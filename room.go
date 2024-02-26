package main

import (
	"errors"
	"io"
	"log/slog"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

type Message map[string]interface{}

type Call struct {
	PeersLock sync.Mutex
	Peers     map[string]*websocket.Conn
}

func (c *Call) AddPeer(token string, ws *websocket.Conn) {
	c.PeersLock.Lock()
	defer c.PeersLock.Unlock()

	const t = "NeedOffer"
	response := struct {
		Type  string
		Token string
	}{t, token}

	delete(c.Peers, token)
	for k, v := range c.Peers {
		slog.Debug("Handler", "type", t, "from", k, "to", token)
		_ = websocket.JSON.Send(v, response)
	}
	c.Peers[token] = ws
}

func (c *Call) RemovePeer(token string) {
	c.PeersLock.Lock()
	defer c.PeersLock.Unlock()

	const t = "Leave"
	response := struct {
		Type  string
		Token string
	}{t, token}

	delete(c.Peers, token)
	for k, v := range c.Peers {
		slog.Debug("Handler", "type", t, "from", token, "to", k)
		_ = websocket.JSON.Send(v, response)
	}
}

var CallsLock sync.Mutex
var Calls map[int]*Call

func GetCall(ID int) *Call {
	CallsLock.Lock()
	call, ok := Calls[ID]
	if !ok {
		call = new(Call)
		call.Peers = make(map[string]*websocket.Conn)
		Calls[ID] = call
		slog.Debug("Started call", "ID", ID)
	}
	CallsLock.Unlock()

	return call
}

func EndCall(ID int) {
	CallsLock.Lock()
	slog.Debug("Ended call", "ID", ID)
	delete(Calls, ID)
	CallsLock.Unlock()
}

func RoomTmplHandler(w http.ResponseWriter, r *http.Request) error {
	id, err := GetIDFromURL(r.URL, "/")
	if err != nil {
		return err
	}
	return WriteTemplateAny(w, "room.tmpl", http.StatusOK, id)
}

func WebsocketRoomHandler(ws *websocket.Conn) {
	if err := func(ws *websocket.Conn) error {
		r := ws.Request()

		id, err := GetIDFromURL(r.URL, "/websocket/")
		if err != nil {
			return err
		}

		token, err := GenerateSessionToken()
		if err != nil {
			return err
		}

		call := GetCall(id)
		call.AddPeer(token, ws)
		defer func() {
			call.RemovePeer(token)

			call.PeersLock.Lock()
			if len(call.Peers) == 0 {
				EndCall(id)
			}
			call.PeersLock.Unlock()
		}()

		var message Message
		for {
			if err := websocket.JSON.Receive(ws, &message); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}

			t, ok := message["Type"].(string)
			if !ok {
				return errors.New("message type is not a string")
			}
			if t == "Ping" {
				continue
			}

			peerToken, ok := message["Token"].(string)
			if !ok {
				return errors.New("token is not a string")
			}

			call.PeersLock.Lock()
			peer, ok := call.Peers[peerToken]
			call.PeersLock.Unlock()
			if ok {
				var response interface{}
				switch t {
				default:
					return errors.New("unrecognized message type")
				case "Answer":
					response = struct {
						Type   string
						Token  string
						Answer interface{}
					}{t, token, message[t]}
				case "ICE":
					response = struct {
						Type  string
						Token string
						ICE   interface{}
					}{t, token, message[t]}
				case "Offer":
					response = struct {
						Type  string
						Token string
						Offer interface{}
					}{t, token, message[t]}
				}
				_ = websocket.JSON.Send(peer, response)
				slog.Debug("Handler", "type", t, "from", token, "to", peerToken)
			}
		}
	}(ws); err != nil {
		slog.Error("Failed to handle websocket connection", "error", err)
	}
	ws.Close()
}
