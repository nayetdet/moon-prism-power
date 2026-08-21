package myanimelist

import (
	"fmt"
	"net"
	"net/http"
	"sync"
)

const callbackAddr = "127.0.0.1:3939"

func startCallbackServer() (<-chan callbackResult, *http.Server, error) {
	callback := make(chan callbackResult, 1)
	server := &http.Server{Addr: callbackAddr, Handler: callbackHandler(callback)}
	listener, err := net.Listen("tcp", server.Addr)
	if err != nil {
		return nil, nil, fmt.Errorf("start %s or configure the callback: %w", callbackAddr, err)
	}

	go func() {
		if serveErr := server.Serve(listener); serveErr != nil && serveErr != http.ErrServerClosed {
			select {
			case callback <- callbackResult{Error: serveErr.Error()}:
			default:
			}
		}
	}()

	return callback, server, nil
}

func callbackHandler(callback chan<- callbackResult) http.HandlerFunc {
	var once sync.Once
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/callback" {
			http.NotFound(w, r)
			return
		}

		once.Do(func() {
			callback <- callbackResult{Code: r.URL.Query().Get("code"), State: r.URL.Query().Get("state"), Error: r.URL.Query().Get("error")}
		})

		_, _ = fmt.Fprint(w, "Authorization received. You can close this page and return to the terminal.")
	}
}
