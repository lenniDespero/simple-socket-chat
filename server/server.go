package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type Message struct {
	Time	time.Time	`json:"time"`
	Nick	string		`json:"nick"`
	Text	string		`json:"text"`
}

type pool struct {
	clients          map[string]*websocket.Conn
	addClientChan    chan *websocket.Conn
	removeClientChan chan *websocket.Conn
	broadcastChan    chan Message
}

var (
	port = flag.String("port", "8888", "Port to connection")
	pass = flag.String("password", "", "Secure connection with password")
)

func main() {
	flag.Parse()
	h := makePool()
	mux := http.NewServeMux()
	mux.Handle("/", websocket.Handler(func(ws *websocket.Conn) {
		handler(ws, h)
	}))
	s := http.Server{Addr: ":" + *port, Handler: mux}

	log.Fatal(s.ListenAndServe())
}

func handler(ws *websocket.Conn, h *pool) {
	go h.run()

	h.addClientChan <- ws

	for {
		var m Message
		err := websocket.JSON.Receive(ws, &m)
		if err != nil {
			h.removeClientChan <- ws
			//delete(h.clients, ws.LocalAddr().String())
			return
		}
		h.broadcastChan <- m
	}
}

func makePool() *pool {
	return &pool{
		clients:          make(map[string]*websocket.Conn),
		addClientChan:    make(chan *websocket.Conn),
		removeClientChan: make(chan *websocket.Conn),
		broadcastChan:    make(chan Message),
	}
}

func (p *pool) run() {
	for {
		select {
		case conn := <-p.addClientChan:
			p.clients[conn.RemoteAddr().String()] = conn
		case conn := <-p.removeClientChan:
			fmt.Println(p.clients)
			delete(p.clients, conn.LocalAddr().String())
			fmt.Println(p.clients)
		case m := <-p.broadcastChan:

			fmt.Println(m)
			for _, conn := range p.clients {
				err := websocket.JSON.Send(conn, m)
				if err != nil {
					fmt.Println(&p.clients, m)
					fmt.Println("Error broadcasting message: ", err)
					return
				}
			}
		}
	}
}

