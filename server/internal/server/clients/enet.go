package clients

import (
	"fmt"
	"log"
	"server/internal/server"
	"server/internal/server/states"
	"server/pkg/packets"

	"github.com/codecat/go-enet"
	"google.golang.org/protobuf/proto"
)

type EnetClient struct {
	id       uint64
	conn     enet.Peer
	hub      *server.Hub
	sendChan chan *packets.Packet
	state    server.ClientStateHandler
	logger   *log.Logger
	dbTx     *server.DbTx
}

func NewEnetClient(conn enet.Peer, hub *server.Hub) *EnetClient {
	return &EnetClient{
		conn:     conn,
		hub:      hub,
		sendChan: make(chan *packets.Packet, 256),
		logger:   log.New(log.Writer(), "Client ", log.LstdFlags),
		dbTx:     hub.NewDbTx(),
	}
}

func (e EnetClient) Initialize(id uint64) {
	e.id = id
	e.logger.SetPrefix(fmt.Sprintf("Client %d: ", e.id))
	e.SetState(&states.Connected{})
}

func (e EnetClient) SetState(newState server.ClientStateHandler) {
	previousStateName := "None"
	if e.state != nil {
		previousStateName = e.state.Name()
		e.state.OnExit()
	}

	newStateName := "None"
	if newState != nil {
		newStateName = newState.Name()
	}

	e.logger.Printf("Changing state from %s to %s", previousStateName, newStateName)
	e.state = newState

	if e.state != nil {
		e.state.SetClient(e)
		e.state.OnEnter()
	}
}

func (e EnetClient) Id() uint64 {
	return e.id
}

func (e EnetClient) ProcessMessage(senderId uint64, msg packets.Msg) {
	e.state.HandleMessage(senderId, msg)
}

func (e EnetClient) SocketSend(message packets.Msg) {
	e.SocketSendAs(message, e.id)
}

func (e EnetClient) SocketSendAs(message packets.Msg, senderId uint64) {
	select {
	case e.sendChan <- &packets.Packet{SenderId: senderId, Msg: message}:
	default:
		e.logger.Printf("Send channel full, dropping message: %T", message)
	}
}

func (e EnetClient) PassToPeer(message packets.Msg, peerId uint64) {
	if peer, exists := e.hub.Clients.Get(peerId); exists {
		peer.ProcessMessage(e.id, message)
	}
}

func (e EnetClient) Broadcast(message packets.Msg) {
	e.hub.BroadcastChan <- &packets.Packet{SenderId: e.id, Msg: message}
}

func (e EnetClient) ReadPump() {
	// Enet does not support reading from a channel, so we just do nothing with this
	// We need to handle the incoming traffic in one place and then send it to each client from there
}

func (e EnetClient) WritePump() {
	defer func() {
		e.logger.Printf("Write pump for client %d stopped", e.id)
		e.Close("write pump stopped")
	}()

	for packet := range e.sendChan {
		if packet == nil {
			continue
		}
		data, err := proto.Marshal(packet)
		if err != nil {
			e.logger.Printf("Failed to marshal packet: %v", err)
			continue
		}

		err = e.conn.SendBytes(data, 0, enet.PacketFlagUnreliableFragment)
		if err != nil {
			e.logger.Printf("Failed to send packet: %v", err)
			continue
		}
	}
}

func (e EnetClient) DbTx() *server.DbTx {
	return e.dbTx
}

func (e EnetClient) SharedGameObjects() *server.SharedGameObjects {
	return e.hub.SharedGameObjects
}

func (e EnetClient) Close(reason string) {
	e.logger.Printf("Closing client connection because: %s", reason)

	e.SetState(nil)

	e.hub.UnregisterChan <- e

	e.conn.Disconnect(uint32(e.id))
	if _, closed := <-e.sendChan; !closed {
		close(e.sendChan)
	}
}
