// +build !js !wasm

package ot

import (
	"encoding/json"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

type Update struct {
	Revision  int       `json:"revision"`
	Operation []Applier `json:"operation"`
}

type Server struct {
	Listeners         []chan []Applier
	RegisterListeners chan chan []Applier
}

func (server *Server) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	conn, _, _, err := ws.UpgradeHTTP(req, res)
	if err != nil {
		log.Print(err)
		return
	}
	defer conn.Close()
	client := Client{
		Socket: conn,
		Send:   make(chan []Applier),
		Brodcast: func(op []Applier) {
			server.BrodcastOperation(op)
		},
	}
	server.RegisterListeners <- client.Send
	go client.Read()
	client.Write()
}

func (server *Server) BrodcastOperation(op []Applier) {
	wg := sync.WaitGroup{}
	wg.Add(len(server.Listeners))
	for li := range server.Listeners {
		if server.Listeners[li] != nil {
			go func() {
				server.Listeners[li] <- op
				wg.Done()
			}()
		} else {
			server.Listeners[li] = server.Listeners[len(server.Listeners)-1]
			server.Listeners[len(server.Listeners)-1] = nil
			server.Listeners = server.Listeners[:len(server.Listeners)-1]
			wg.Done()
		}
	}
	wg.Wait()
}

type Client struct {
	Send     chan []Applier
	Socket   net.Conn
	Brodcast func([]Applier)
	Done     chan struct{}
}

func (client *Client) Read() {
	for {
		var op []Applier
		bts, _, err := wsutil.ReadClientData(client.Socket)
		if err != nil {
			log.Print(err)
			break
		}
		if err := json.Unmarshal(bts, &op); err != nil {
			log.Print(err)
			break
		}
		log.Printf("%v", op)
		client.Brodcast(op)
	}
	close(client.Send)
}

func (client *Client) Write() {
	defer client.Socket.Close()
	for op := range client.Send {
		bts, err := json.Marshal(op)
		if err != nil {
			log.Print(err)
			break
		}
		if err := wsutil.WriteServerText(client.Socket, bts); err != nil {
			log.Print(err)
			break
		}
	}
	client.Send = nil
}
