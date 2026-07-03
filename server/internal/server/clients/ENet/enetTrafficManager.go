package ENet

import (
	"fmt"
	"log"
	"server/internal"
	"server/internal/server"
	"server/internal/server/clients"
	"server/pkg/packets"

	"github.com/codecat/go-enet"
	"google.golang.org/protobuf/proto"
)

func InitEnet(h *server.Hub) {
	enet.Initialize()
	host, err := enet.NewHost(enet.NewAddress("127.0.0.1", internal.ENetPort), internal.EnetMaxClients, internal.ENetChannels, 0, 0)
	if err != nil {
		log.Fatalf("Failed to initialize enet: %v", err.Error())
		return
	}

	log.Println("Enet initialized on 8081")
	go TrafficHandler(host, h)
}
func Serve(getNewClient func(conn enet.Peer, h *server.Hub) *clients.EnetClient, conn enet.Peer, h *server.Hub) *clients.EnetClient {
	client := getNewClient(conn, h)

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
		switch ev.GetType() {
		case enet.EventConnect:
			c := Serve(clients.NewEnetClient, ev.GetPeer(), h)
			c.Initialize(uint64(ev.GetData()))
			cs, _ := h.Clients.Get(uint64(ev.GetData()))
			cs.ENet = c
			c.SetClient(cs)

			h.PeerToClient[ev.GetPeer().GetConnectId()] = c.Client()
		case enet.EventDisconnect:
			id := ev.GetData()
			client, _ := h.Clients.Get(uint64(id))
			if client != nil {
				h.UnregisterChan <- client.ENet
			}
		case enet.EventReceive:
			raw := ev.GetPacket().GetData()

			if len(raw) < 12+16 { // nonce + tag minimum
				log.Printf("Packet too short")
				continue
			}

			client, _ := h.PeerToClient[ev.GetPeer().GetConnectId()]
			if client == nil {
				log.Printf("Unknown client")
				continue
			}

			// split the packet
			nonce := raw[:12]
			ciphertext := raw[12 : len(raw)-16]
			tag := raw[len(raw)-16:]

			// decrypt: append tag to ciphertext because chacha20poly1305.Open expects it
			encData := append(ciphertext, tag...)

			plaintext, err := client.Cyph.Open(nil, nonce, encData, nil)
			if err != nil {
				log.Printf("Failed to decrypt packet from client %d: %v", client.Id(), err)
				continue
			}

			// unmarshal the protobuf from decrypted payload
			packet := &packets.Packet{}
			err = proto.Unmarshal(plaintext, packet)
			if err != nil {
				log.Printf("Failed to unmarshal packet: %v", err)
				continue
			}

			go client.ENet.ProcessMessage(packet.SenderId, packet.Msg, int(ev.GetChannelID()))
		default:
			panic("unhandled default case")
		}

	}
}
