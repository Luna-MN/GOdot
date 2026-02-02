package states

import (
	"context"
	"fmt"
	"log"
	"server/internal/server"
	"server/internal/server/db"
	"server/internal/server/objects"
	"server/pkg/packets"
	"time"
)

type inGame struct {
	client                 server.ClientInterfacer
	player                 *objects.Player
	logger                 *log.Logger
	cancelPlayerUpdateLoop context.CancelFunc
	queries                *db.Queries
	dbCtx                  context.Context
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
	g.player.Speed = 25

	// Send the initial state of the player to the client
	g.client.SocketSend(packets.NewPlayer(g.client.Id(), g.player))
}

func (g *inGame) HandleMessage(senderId uint64, message packets.Msg) {
	switch message := message.(type) {
	case *packets.Packet_Player:
		g.handlePlayer(senderId, message)
	case *packets.Packet_PlayerInput:
		g.handlePlayerInput(senderId, message)
	}
}

func (g *inGame) syncPlayer(delta float32) {
	newX := g.player.Position.X + g.player.Speed*g.player.Direction.X*delta
	newY := g.player.Position.Y + g.player.Speed*g.player.Direction.Y*delta
	newZ := g.player.Position.Z + g.player.Speed*g.player.Direction.Z*delta
	//g.logger.Printf("Player %d moving to (%f, %f, %f)", g.client.Id(), newX, newY, newZ)
	g.player.Position = objects.NewVector3(newX, newY, newZ)

	updatePacket := packets.NewPlayer(g.client.Id(), g.player)
	g.client.Broadcast(updatePacket)
	go g.client.SocketSend(updatePacket)
}

func (g *inGame) HandleChatMessage(id uint64, message *packets.Packet_Chat) {

}

func (g *inGame) OnExit() {
	g.logger.Printf("Client %d exited game state", g.client.Id())
	if g.cancelPlayerUpdateLoop != nil {
		g.cancelPlayerUpdateLoop()
	}
	g.client.SharedGameObjects().Players.Remove(g.client.Id())
}

func (g *inGame) updatePlayerLoop(ctx context.Context) {
	const delta float32 = 0.005
	ticker := time.NewTicker(time.Duration(delta * float32(time.Second)))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			g.syncPlayer(delta)
		case <-ctx.Done():
			return
		}
	}
}

func (g *inGame) handlePlayer(id uint64, message *packets.Packet_Player) {
	if id == g.client.Id() {
		g.logger.Printf("Received player update from our own client, ignoring (ID %d)", id)
		return
	}

	g.client.SocketSendAs(message, id)
}

func (g *inGame) handlePlayerInput(id uint64, message *packets.Packet_PlayerInput) {
	if id == g.client.Id() {
		g.player.Direction = objects.NewVector3(message.PlayerInput.X, message.PlayerInput.Y, message.PlayerInput.Z)
		g.logger.Printf("Player %d input: %v", g.client.Id(), g.player.Direction)
		// if this is the first input, start the player update loop
		if g.cancelPlayerUpdateLoop == nil {
			ctx, cancel := context.WithCancel(context.Background())
			g.cancelPlayerUpdateLoop = cancel
			go g.updatePlayerLoop(ctx)
		}
	}
}
