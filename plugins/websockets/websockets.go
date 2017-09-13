package websockets

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/codeamp/transistor"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

type Websockets struct {
	ServiceAddress string `mapstructure:"service_address"`
	Events         chan transistor.Event
	clients        map[*Client]bool
	broadcast      chan []byte
	register       chan *Client
	unregister     chan *Client
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewWebsockets() *Websockets {
	return &Websockets{}
}

func init() {
	transistor.RegisterPlugin("websockets", func() transistor.Plugin { return NewWebsockets() })
}

func (x *Websockets) Listen() {
	go x.run()

	r := mux.NewRouter()
	r.HandleFunc("/", x.serveWs)

	err := http.ListenAndServe(fmt.Sprintf("%s", x.ServiceAddress), r)
	if err != nil {
		log.Printf("Error starting server: %v", err)
	}
}

// serveWs handles websocket requests from the peer.
func (x *Websockets) serveWs(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	client := &Client{ws: x, conn: conn, send: make(chan []byte, 256)}
	client.ws.register <- client
	go client.writePump()
	client.readPump()
}

func (x *Websockets) Start(e chan transistor.Event) error {
	x.Events = e
	x.broadcast = make(chan []byte)
	x.register = make(chan *Client)
	x.unregister = make(chan *Client)
	x.clients = make(map[*Client]bool)

	go x.Listen()
	log.Printf("Started the Websockets service on %s\n", x.ServiceAddress)

	return nil
}

func (x *Websockets) Stop() {
	log.Println("Stopping Websockets")
}

func (x *Websockets) Subscribe() []string {
	return []string{
		"plugins.WebsocketMsg",
	}
}

func (x *Websockets) Process(e transistor.Event) error {
	log.Printf("Process Websockets event: %s", e.Name)
	if e.Name == "plugins.WebsocketMsg" {
		json, _ := json.Marshal(e.Payload)
		x.broadcast <- json
	}
	return nil
}

func (x *Websockets) run() {
	for {
		select {
		case client := <-x.register:
			x.clients[client] = true
		case client := <-x.unregister:
			if _, ok := x.clients[client]; ok {
				delete(x.clients, client)
				close(client.send)
			}
		case message := <-x.broadcast:
			for client := range x.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(x.clients, client)
				}
			}
		}
	}
}
