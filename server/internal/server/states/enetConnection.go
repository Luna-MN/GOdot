package states

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"server/internal/server"
	"server/internal/server/db"
	"server/internal/server/objects"
	"server/pkg/packets"
	"time"

	"golang.org/x/crypto/chacha20poly1305"
)

type EnetConnection struct {
	client                 *server.Client
	player                 *objects.Player
	logger                 *log.Logger
	cancelPlayerUpdateLoop context.CancelFunc
	queries                *db.Queries
	dbCtx                  context.Context
	Hub                    *server.Hub
}

func (e *EnetConnection) CanReceiveGlobalChat() bool {
	return false
}

func (e *EnetConnection) Name() string {
	return "EnetConnection"
}

func (e *EnetConnection) SetClient(client *server.Client) {
	e.client = client
	loginPrefix := fmt.Sprintf("Client %d [%s]: ", client.Id(), e.Name())
	e.logger = log.New(log.Writer(), loginPrefix, log.LstdFlags)
	e.queries = client.DbTx().Queries
	e.dbCtx = client.DbTx().Ctx
	key := make([]byte, 32)
	rand.Read(key)
	cypher, err := chacha20poly1305.New(key)
	if err != nil {
		e.logger.Fatal(err)
	}

	e.client.SetCyph(cypher)

	var keyMsg = packets.NewKey(e.client.Id(), key, time.Now().Add(time.Hour).Unix())

	e.client.SocketSend(keyMsg, server.WebSocket)
}

func (e *EnetConnection) OnEnter() {
	e.logger.Printf("Sending ID to client")
}

func (e *EnetConnection) HandleMessage(senderId uint64, message packets.Msg, transfer server.TransferType) {
}

func (e *EnetConnection) OnExit() {

}
