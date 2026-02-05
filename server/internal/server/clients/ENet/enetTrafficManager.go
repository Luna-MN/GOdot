package ENet

import (
	"fmt"
	"log"
	"server/internal/server"
	"server/internal/server/clients"
	"server/pkg/packets"
	"strconv"

	"github.com/codecat/go-enet"
	"google.golang.org/protobuf/proto"
)

func InitEnet(h *server.Hub) {
	enet.Initialize()
	host, err := enet.NewHost(enet.NewAddress("127.0.0.1", 8081), 4000, 3, 0, 0)
	if err != nil {
		log.Fatalf("Failed to initialize enet: %v", err.Error())
		return
	}

	log.Println("Enet initialized on 8081")
	go TrafficHandler(host, h)
}
func Serve(getNewClient func(conn enet.Peer, h *server.Hub) *clients.EnetClient, conn enet.Peer, h *server.Hub) *clients.EnetClient {
	client := getNewClient(conn, h)

	h.RegisterChan <- client

	go client.WritePump()

	return client
}

func TrafficHandler(host enet.Host, h *server.Hub) {
	defer func() {
		fmt.Println("Enet traffic handler exiting")
		host.Destroy()
	}()
	for {
		ev := host.Service(10)
		if ev.GetType() == enet.EventNone {
			continue
		}
		fmt.Println(ev.GetType())
		switch ev.GetType() {
		case enet.EventConnect:
			fmt.Println("New Enet connection " + strconv.Itoa(int(ev.GetData())))
			c := Serve(clients.NewEnetClient, ev.GetPeer(), h)
			cs, _ := h.Clients.Get(uint64(ev.GetData()))
			cs.ENet = c
			cs.WS.SetEnet(c)
			c.Initialize(uint64(int(ev.GetData())))
		case enet.EventDisconnect:
			client, _ := h.Clients.Get(uint64(ev.GetPeer().GetConnectId()))
			h.UnregisterChan <- client.ENet
		case enet.EventReceive:
			client, _ := h.Clients.Get(uint64(ev.GetPeer().GetConnectId()))
			packet := &packets.Packet{}
			err := proto.Unmarshal(ev.GetPacket().GetData(), packet)
			if err != nil {
				log.Printf("Failed to unmarshal packet: %v", err)
				continue
			}

			client.ENet.ProcessMessage(packet.SenderId, packet.Msg)
		default:
			panic("unhandled default case")
		}

	}
}
