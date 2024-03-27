package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/net/websocket"
)

type RequestHandler func(w http.ResponseWriter, r *http.Request) error

var DebugMode = "off"

func (fn RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	statusCode := http.StatusOK
	level := slog.LevelDebug

	start := time.Now()

	err := fn(w, r)
	if err != nil {
		var httpError HTTPError
		displayError := err
		if (errors.As(err, &httpError)) && (httpError.StatusCode != http.StatusInternalServerError) {
			statusCode = httpError.StatusCode
			level = slog.LevelWarn
		} else {
			statusCode = http.StatusInternalServerError
			if DebugMode != "on" {
				displayError = TryAgainLaterError
			}
			level = slog.LevelError
		}

		if err := WriteTemplate(w, "error.tmpl", statusCode, nil, displayError); err != nil {
			defer slog.Error("Failed to write error template", "error", err)
		}
	}

	slog.LogAttrs(context.Background(), level, "Request from", slog.String("client", r.RemoteAddr), slog.String("method", r.Method), slog.String("path", r.URL.Path), slog.Int("status", statusCode), slog.Duration("duration", time.Now().Sub(start)), slog.Any("error", err))
}

func GetRouter() *http.ServeMux {
	mux := http.NewServeMux()
	ParseTemplates()

	mux.Handle("/", RequestHandler(RoomTmplHandler))
	mux.Handle("/favicon.ico", http.NotFoundHandler())
	mux.Handle("/ws/", websocket.Handler(WebsocketRoomHandler))

	return mux
}

func main() {
	if DebugMode == "on" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	}

	port := flag.Uint("p", 7071, "port value to listen on")
	flag.Parse()
	netAddr := fmt.Sprintf("0.0.0.0:%d", *port)

	var server *http.Server

	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	go func() {
		signal := <-sigchan
		slog.Info("Received", "signal", signal)
		slog.Info("Call is exitting...")

		server.Close()
		os.Exit(0)
	}()

	ready := make(chan struct{})
	tmplsModified := make(chan struct{})
	go MonitorTemplates(ready, tmplsModified)
	<-ready

	Calls = make(map[int]*Call)

	for {
		server = &http.Server{Addr: netAddr, Handler: GetRouter()}
		go func() {
			slog.Info("Listening on", "addr", netAddr)
			if err := server.ListenAndServe(); err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()
		<-tmplsModified
		slog.Info("Detected template file changes. Reloading...")
		server.Close()
	}
}
