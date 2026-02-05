package server

import (
	"context"
	"database/sql"
	_ "embed"
	"log"
	"net/http"
	"server/internal/server/db"
	"server/internal/server/objects"
	"server/pkg/packets"

	_ "modernc.org/sqlite"
)

//go:embed db/config/schema.sql
var schemaGenSql string

// DbTx A structure for database transaction context
type DbTx struct {
	Ctx     context.Context
	Queries *db.Queries
}

func (h *Hub) NewDbTx() *DbTx {
	return &DbTx{
		Ctx:     context.Background(),
		Queries: db.New(h.dbPool),
	}
}

type SharedGameObjects struct {
	// ID of the player is the ID of the client
	Players *objects.SharedCollection[*objects.Player]
}

// ClientStateHandler A structure for a state macgine to process the client's messages
type ClientStateHandler interface {
	Name() string

	// SetClient Inject thje client into the state machine
	SetClient(client ClientInterfacer)

	OnEnter()
	HandleMessage(senderId uint64, message packets.Msg)
	// OnExit clean up the state machine
	OnExit()
}
type Client struct {
	ClientId uint64
	ENet     ClientInterfacer
	WS       ClientInterfacer
}

func NewClient(WS ClientInterfacer) Client {
	return Client{
		ClientId: 0,
		ENet:     nil,
		WS:       WS,
	}
}

type ClientInterfacer interface {
	// Initialize sets the client's Id and anything else to be initialized
	Initialize(id uint64)
	SetState(newState ClientStateHandler)
	Id() uint64
	Enet() ClientInterfacer
	SetEnet(Enet ClientInterfacer)
	WS() ClientInterfacer
	SetWS(WS ClientInterfacer)
	ProcessMessage(senderId uint64, msg packets.Msg)

	// SocketSend Puts data from this client into the write pump
	SocketSend(message packets.Msg)

	// SocketSendAs puts data from another client into the write pump
	SocketSendAs(message packets.Msg, senderId uint64)

	// PassToPeer forward message to another client for processing
	PassToPeer(message packets.Msg, peerId uint64)

	// Broadcast send message to all clients except the sender
	Broadcast(message packets.Msg)

	// ReadPump Pump data from the connected socket directly to the client
	ReadPump()

	// WritePump Pump data from the client directly to the connected socket
	WritePump()

	// DbTx A reference to the database transation context for this client
	DbTx() *DbTx

	SharedGameObjects() *SharedGameObjects

	// Close closes the connection
	Close(reason string)
}

// Hub is the central point of communication for the server, managing connections and broadcasting messages to clients.
type Hub struct {
	Clients         *objects.SharedCollection[Client]
	HomelessClients *objects.SharedCollection[ClientInterfacer]
	//Clients *objects.SharedCollection[ClientInterfacer]

	// Packets in this channel will be processed by all connected clients except the sender
	BroadcastChan chan *packets.Packet

	// Clients that want to register
	RegisterChan chan ClientInterfacer

	// Clients that want to unregister
	UnregisterChan chan ClientInterfacer

	dbPool *sql.DB

	SharedGameObjects *SharedGameObjects
}

func NewHub() *Hub {
	dbPool, err := sql.Open("sqlite", "db.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	return &Hub{
		Clients:         objects.NewSharedCollection[Client](),
		HomelessClients: objects.NewSharedCollection[ClientInterfacer](),
		BroadcastChan:   make(chan *packets.Packet),
		RegisterChan:    make(chan ClientInterfacer),
		UnregisterChan:  make(chan ClientInterfacer),
		dbPool:          dbPool,
		SharedGameObjects: &SharedGameObjects{
			Players: objects.NewSharedCollection[*objects.Player](),
		},
	}
}

func (h *Hub) Run() {
	log.Println("Initializing database")
	_, err := h.dbPool.ExecContext(context.Background(), schemaGenSql)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Awaiting clients registration")
	for {
		select {
		case client := <-h.RegisterChan:
			var clients = NewClient(client)
			var Id = h.Clients.Add(clients)
			clients.ClientId = Id
			client.Initialize(Id)
		case client := <-h.UnregisterChan:
			h.Clients.Remove(client.Id())
		case packet := <-h.BroadcastChan:
			//for id, client := range h.Clients {
			h.Clients.ForEach(func(id uint64, client Client) {
				if id != packet.SenderId {
					client.WS.ProcessMessage(packet.SenderId, packet.Msg)
					if client.ENet != nil {
						client.ENet.ProcessMessage(packet.SenderId, packet.Msg)
					}

				}
			})
		}
	}
}

func (h *Hub) Serve(getNewClient func(*Hub, http.ResponseWriter, *http.Request) (ClientInterfacer, error), writer http.ResponseWriter, request *http.Request) {
	log.Println("New client connected from", request.RemoteAddr)
	client, err := getNewClient(h, writer, request)

	if err != nil {
		log.Printf("Error obtaining client for new connection: %v", err)
		return
	}

	h.RegisterChan <- client

	go client.WritePump()
	go client.ReadPump()
}
