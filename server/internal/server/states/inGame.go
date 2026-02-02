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

type inGame struct {
	client  server.ClientInterfacer
	player  *objects.Player
	logger  *log.Logger
	queries *db.Queries
	dbCtx   context.Context
}

func (g *inGame) Name() string {
	return "InGame"
}

func (g *inGame) SetClient(client server.ClientInterfacer) {
	g.client = client
	loginPrefix := fmt.Sprintf("Client %d [%s]: ", client.Id(), g.Name())
	g.logger = log.New(log.Writer(), loginPrefix, log.LstdFlags)
	g.queries = client.DbTx().Queries
	g.dbCtx = client.DbTx().Ctx
}

func (g *inGame) OnEnter() {
	g.logger.Printf("Client %d entered game state", g.client.Id())
	go g.client.SharedGameObjects().Players.Add(g.player, g.client.Id())

	// Set the initial state of the player
	g.player.Position = objects.NewVector3(0, 0, 0)
	g.player.Speed = 150

	// Send the initial state of the player to the client
	g.client.SocketSend(packets.NewPlayer(g.client.Id(), g.player))
}

func (g *inGame) HandleMessage(senderId uint64, message packets.Msg) {
	switch message := message.(type) {
	case *packets.Packet_Chat:
		g.HandleChatMessage(senderId, message)
	}
}

func (g *inGame) HandleChatMessage(id uint64, message *packets.Packet_Chat) {

}

func (g *inGame) OnExit() {
	g.logger.Printf("Client %d exited game state", g.client.Id())
	g.client.SharedGameObjects().Players.Remove(g.client.Id())
}
