package states

import (
	"context"
	"fmt"
	"log"
	"server/internal/server"
	"server/internal/server/db"
	"server/internal/server/objects"
	"server/pkg/packets"
)

type EnetConnection struct {
	client                 server.ClientInterfacer
	player                 *objects.Player
	logger                 *log.Logger
	cancelPlayerUpdateLoop context.CancelFunc
	queries                *db.Queries
	dbCtx                  context.Context
	hub                    *server.Hub
}

func (e *EnetConnection) Name() string {
	return "EnetConnection"
}

func (e *EnetConnection) SetClient(client server.ClientInterfacer) {
	e.client = client
	loginPrefix := fmt.Sprintf("Client %d [%s]: ", client.Id(), e.Name())
	e.logger = log.New(log.Writer(), loginPrefix, log.LstdFlags)
	e.queries = client.DbTx().Queries
	e.dbCtx = client.DbTx().Ctx
	e.client.SocketSend(packets.NewId(e.client.Id()))
}

func (e *EnetConnection) OnEnter() {
	e.logger.Printf("Sending ID to client")
	e.client.SocketSend(packets.NewId(e.client.Id()))
}

func (e *EnetConnection) HandleMessage(senderId uint64, message packets.Msg) {
	if senderId == e.client.Id() {
		e.client.SetState(&Connected{})
	}
}

func (e *EnetConnection) OnExit() {

}
