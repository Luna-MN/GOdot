package clients

import (
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"log"
	"server/internal/server"
	"server/pkg/packets"

	"github.com/codecat/go-enet"
	"google.golang.org/protobuf/proto"
)

type EnetPacket struct {
	packet  *packets.Packet
	channel int
}

func NewEnetPacket(packet *packets.Packet, channel int) *EnetPacket {
	return &EnetPacket{packet: packet, channel: channel}
}

type EnetClient struct {
	id       uint64
	conn     enet.Peer
	hub      *server.Hub
	sendChan chan *EnetPacket
	logger   *log.Logger
	dbTx     *server.DbTx
	WSConn   *server.ClientInterfacer
	client   *server.Client
}

func (e *EnetClient) Client() *server.Client {
	return e.client
}

func (e *EnetClient) SetClient(client *server.Client) {
	e.client = client
}

func (e *EnetClient) IsEnet() bool {
	return true
}

func NewEnetClient(conn enet.Peer, hub *server.Hub) *EnetClient {
	return &EnetClient{
		conn:     conn,
		hub:      hub,
		sendChan: make(chan *EnetPacket, 256),
		logger:   log.New(log.Writer(), "ENET Client ", log.LstdFlags),
		dbTx:     hub.NewDbTx(),
	}
}

func (e *EnetClient) Initialize(id uint64) {
	e.id = id
	e.logger.SetPrefix(fmt.Sprintf("ENET Client %d: ", e.id))
}

func (e *EnetClient) SetState(newState server.ClientStateHandler, sync ...bool) {
	e.client.SetState(newState)
}

func (e *EnetClient) Id() uint64 {
	return e.id
}

func (e *EnetClient) ProcessMessage(senderId uint64, msg packets.Msg, enetChannel ...int) {
	chnl := server.ENet_Channel0
	if enetChannel == nil {
		return
	} else {
		if enetChannel[0] == 2 {
			chnl = server.ENet_Channel1
		}
	}
	e.client.ProcessMessage(senderId, msg, chnl)
}

func (e *EnetClient) SocketSend(message packets.Msg, enetChannel ...int) {
	e.SocketSendAs(message, e.id, enetChannel...)
}

func (e *EnetClient) SocketSendAs(message packets.Msg, senderId uint64, enetChannel ...int) {
	select {
	case e.sendChan <- NewEnetPacket(&packets.Packet{SenderId: senderId, Msg: message}, enetChannel[0]):
	default:
		e.logger.Printf("Send channel full, dropping message: %T", message)
	}
}

func (e *EnetClient) PassToPeer(message packets.Msg, peerId uint64) {
	if peer, exists := e.hub.Clients.Get(peerId); exists {
		peer.ENet.ProcessMessage(e.id, message)
	}
}

func (e *EnetClient) Broadcast(message packets.Msg, enet bool) {
	e.hub.BroadcastChan <- &server.BroadcastItem{Packet: &packets.Packet{SenderId: e.id, Msg: message}, Enet: enet}
}

func (e *EnetClient) ReadPump() {
	// Enet does not support reading from a channel, so we just do nothing with this
	// We need to handle the incoming traffic in one place and then send it to each client from there
}

func (e *EnetClient) WritePump() {
	defer func() {
		e.logger.Printf("Write pump for client %d stopped", e.id)
		e.Close("write pump stopped")
	}()

	for packet := range e.sendChan {
		if packet == nil {
			continue
		}
		data, err := proto.Marshal(packet.packet)
		if err != nil {
			e.logger.Printf("Failed to marshal packet: %v", err)
			continue
		}
		// nonce ensembles the packet ID with a random number
		nonce := make([]byte, 12)
		binary.LittleEndian.PutUint64(nonce[4:], e.hub.PacketCounter)
		rand.Read(nonce[:4])
		e.hub.PacketCounter++

		var encData = e.client.Cyph.Seal(nil, nonce, data, nil)

		var sendData = append(nonce, encData...)

		err = e.conn.SendBytes(sendData, uint8(packet.channel), enet.PacketFlagUnreliableFragment)
		if err != nil {
			e.logger.Printf("Failed to send packet: %v", err)
			continue
		}
	}
}

func (e *EnetClient) DbTx() *server.DbTx {
	return e.dbTx
}

func (e *EnetClient) SharedGameObjects() *server.SharedGameObjects {
	return e.hub.SharedGameObjects
}

func (e *EnetClient) Close(reason string) {
	e.logger.Printf("Closing client connection because: %s", reason)

	e.SetState(nil)

	e.hub.UnregisterChan <- e

	e.conn.Disconnect(uint32(e.id))
	if _, closed := <-e.sendChan; !closed {
		close(e.sendChan)
	}
}
