package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"server/pkg/packets"

	"server/internal/server"
	"server/internal/server/clients"

	"github.com/codecat/go-enet"
	"google.golang.org/protobuf/proto"
)

var (
	port = flag.Int("port", 8080, "port to run server on")
)

func main() {
	flag.Parse()

	// Define the game hub
	hub := server.NewHub()

	// define the handler for the websocket connections
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		hub.Serve(clients.NewWebSocketClient, w, r)
	})
	initEnet(hub)
	go hub.Run()
	addr := fmt.Sprintf(":%d", *port)

	log.Printf("Server started on %s", addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func initEnet(h *server.Hub) {
	host, err := enet.NewHost(enet.NewListenAddress(8081), 4000, 3, 0, 0)
	if err != nil {
		log.Fatalf("Failed to initialize enet: %v", err.Error())
		return
	}
	go EnetTrafficHandler(host, h)
}
func EnetServe(getNewClient func(conn enet.Peer, h *server.Hub) *clients.EnetClient, conn enet.Peer, h *server.Hub) *clients.EnetClient {
	client := getNewClient(conn, h)

	h.RegisterChan <- client

	go client.WritePump()

	return client
}

func EnetTrafficHandler(host enet.Host, h *server.Hub) {
	defer func() {
		// cleanup
	}()
	for {
		ev := host.Service(1000)
		if ev.GetType() == enet.EventNone {
			continue
		}

		switch ev.GetType() {
		case enet.EventConnect:
			id := h.Clients.Add(EnetServe(clients.NewEnetClient, ev.GetPeer(), h), uint64(ev.GetPeer().GetConnectId()))
			c, _ := h.Clients.Get(id)
			c.Initialize(id)
		case enet.EventDisconnect:
			client, _ := h.Clients.Get(uint64(ev.GetPeer().GetConnectId()))
			h.UnregisterChan <- client
		case enet.EventReceive:
			client, _ := h.Clients.Get(uint64(ev.GetPeer().GetConnectId()))
			packet := &packets.Packet{}
			err := proto.Unmarshal(ev.GetPacket().GetData(), packet)
			if err != nil {
				log.Printf("Failed to unmarshal packet: %v", err)
				continue
			}

			client.ProcessMessage(packet.SenderId, packet.Msg)
		default:
			panic("unhandled default case")
		}

	}

}
