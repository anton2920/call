//go:build freebsd

package main

import (
	"io/fs"
	"log"
	"log/slog"
	"os"
	"syscall"
	"time"
	"unsafe"
)

func MonitorTemplates(ready, tmplsModified chan<- struct{}) {
	var event syscall.Kevent_t

	kq, err := syscall.Kqueue()
	if err != nil {
		log.Fatal("Failed to open a kernel queue: ", err)
	}

	fs.WalkDir(os.DirFS(TemplatesDirName), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Fatal("Failed to walk directory: ", err)
		}

		if !d.IsDir() {
			fname := TemplatesDirName + string(os.PathSeparator) + path
			fd, err := syscall.Open(fname, os.O_RDONLY, 0644)
			if err != nil {
				return WrapErrorWithTrace(err)
			}

			event := syscall.Kevent_t{Ident: uint64(fd), Filter: syscall.EVFILT_VNODE, Flags: syscall.EV_ADD | syscall.EV_CLEAR, Fflags: syscall.NOTE_WRITE}
			if _, err := syscall.Kevent(kq, unsafe.Slice(&event, 1), nil, nil); err != nil {
				return WrapErrorWithTrace(err)
			}
		}

		return nil
	})

	close(ready)
	for {
		if _, err := syscall.Kevent(kq, nil, unsafe.Slice(&event, 1), nil); err != nil {
			if err != syscall.EINTR {
				slog.Error("Failed to get kernel events", "error", err)
			}
			continue
		}

		tmplsModified <- struct{}{}

		/* Sleeping to prevent runaway events. */
		time.Sleep(time.Millisecond * 100)

	}
}
