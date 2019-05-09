// +build !js !wasm

package ot

import (
	"encoding/json"
	"fmt"
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

type ApplierList []Applier

func (list *ApplierList) UnmarshalJSON(data []byte) error {
	var marshaled []interface{}
	if err := json.Unmarshal(data, &marshaled); err != nil {
		return err
	}
	li := ApplierList(make([]Applier, len(marshaled)))
	list = &li
	for i, op := range marshaled {
		switch o := op.(type) {
		case string:
			(*list)[i] = Insert(o)
		case float64:
			if o < 0 {
				(*list)[i] = Delete(int(o))
			} else {
				(*list)[i] = Retain(int(o))
			}
		default:
			return fmt.Errorf("unknown op type %v %t", o, o)
		}
	}
	return nil
}
