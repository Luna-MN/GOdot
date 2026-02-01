package states

import (
	"context"
	"fmt"
	"log"
	"server/internal/server"
	"server/internal/server/db"
	"server/pkg/packets"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type Connected struct {
	client  server.ClientInterfacer
	logger  *log.Logger
	queries *db.Queries
	dbCtx   context.Context
}

func (c *Connected) Name() string {
	return "Connected"
}

func (c *Connected) SetClient(client server.ClientInterfacer) {
	c.client = client
	loginPrefix := fmt.Sprintf("Client %d [%s]: ", client.Id(), c.Name())
	c.logger = log.New(log.Writer(), loginPrefix, log.LstdFlags)
	c.queries = client.DbTx().Queries
	c.dbCtx = client.DbTx().Ctx
}

func (c *Connected) OnEnter() {
	c.client.SocketSend(packets.NewId(c.client.Id()))
}

func (c *Connected) HandleMessage(senderId uint64, message packets.Msg) {
	switch message := message.(type) {
	case *packets.Packet_LoginRequest:
		c.HandleLoginRequest(senderId, message)
	case *packets.Packet_RegisterRequest:
		c.HandleRegisterRequest(senderId, message)
	}
}

func (c *Connected) OnExit() {
}

func (c *Connected) HandleLoginRequest(id uint64, message *packets.Packet_LoginRequest) {
	if id != c.client.Id() {
		c.logger.Printf("Received login request from differant client, ignoring (ID %d)", id)
		return
	}

	username := message.LoginRequest.Username

	genericFailMessage := packets.NewDeny("Incorrect username or password")

	user, err := c.queries.GetUserByUsername(c.dbCtx, strings.ToLower(username))
	if err != nil {
		c.logger.Printf("Error getting user by username: %v", err)
		c.client.SocketSend(genericFailMessage)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(message.LoginRequest.Password))
	if err != nil {
		c.logger.Printf("User Password incorrect: %v", err)
		c.client.SocketSend(genericFailMessage)
		return
	}

	c.logger.Printf("Login successful for user %s", username)
	c.client.SocketSend(packets.NewOK())
}

func (c *Connected) HandleRegisterRequest(id uint64, message *packets.Packet_RegisterRequest) {
	if id != c.client.Id() {
		c.logger.Printf("Received register request from differant client, ignoring (ID %d)", id)
		return
	}

	username := message.RegisterRequest.Username
	err := validateUsername(username)
	if err != nil {
		c.logger.Printf("Invalid username: %v", err)
		c.client.SocketSend(packets.NewDeny(err.Error()))
		return
	}

	if _, err = c.queries.GetUserByUsername(c.dbCtx, strings.ToLower(username)); err != nil {
		c.logger.Printf("Username already taken: %v", err)
		c.client.SocketSend(packets.NewDeny("Username already taken"))
		return
	}

	genericFailMessage := packets.NewDeny("Error registering user (internal server error)")

	hash, err := bcrypt.GenerateFromPassword([]byte(message.RegisterRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		c.logger.Printf("Error hashing password: %v", err)
		c.client.SocketSend(genericFailMessage)
		return
	}

	_, err = c.queries.CreateUser(c.dbCtx, db.CreateUserParams{
		Username:     strings.ToLower(username),
		PasswordHash: string(hash),
	})
	if err != nil {
		c.logger.Printf("Error creating user: %v", err)
		c.client.SocketSend(genericFailMessage)
		return
	}

	c.logger.Printf("User %s registered successfully", username)
	c.client.SocketSend(packets.NewOK())
}

func validateUsername(username string) error {

	if len(username) < 3 || len(username) > 20 {
		return fmt.Errorf("Username must be between 3 and 20 characters long")
	}

	return nil
}
