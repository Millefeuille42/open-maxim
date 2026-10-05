package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maxim/internal/events"
	"maxim/internal/events/local"
	"maxim/internal/models"
	"maxim/internal/state"
	"maxim/internal/state/memory"
	"maxim/internal/utils"
	"net"
	"sync"
	"time"
)

type Server struct {
	Port int
	Hash *utils.Argon2idHash
	Salt string

	Clients         map[*events.Client]bool
	clientWaitGroup sync.WaitGroup

	Users    state.Engine[*models.User]
	Channels state.Engine[*models.Channel]

	Events events.Engine

	newConn    chan net.Conn
	disconnect chan *events.Client

	listener net.Listener
}

func NewServer(port int, salt string) *Server {
	return &Server{
		Port:    port,
		Hash:    utils.NewArgon2idHash(1, 32, 64*1024, 32, 256),
		Salt:    salt,
		Clients: make(map[*events.Client]bool),

		// TODO add permanent state driver like a DB?
		Users:    memory.NewEngine[*models.User](),
		Channels: memory.NewEngine[*models.Channel](),

		// TODO: Handle this as argument
		Events: local.NewLocalEngine(100),

		newConn:    make(chan net.Conn),
		disconnect: make(chan *events.Client),
	}
}

func (s *Server) handleConnections() {
	log.Printf("Server is listening on port %d\n", s.Port)
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Println(err)
			time.Sleep(5 * time.Millisecond)
			continue
		}
		s.newConn <- conn
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	client := events.NewClient(conn, s.disconnect)
	client.Run(s.Events)
	s.Clients[client] = true
}

func (s *Server) commandRouter(msg events.Message, ctx context.Context) {
	if handler, ok := handlers[msg.Command]; ok {
		handler(s, msg, ctx)
		return
	}
	log.Printf("Unknown command: %s", msg.Command)
}

func (s *Server) removeClient(client *events.Client) {
	if _, ok := s.Users.Get(client.Username); !ok {
		s.Users.Remove(client.Username)
	}
	delete(s.Clients, client)
	s.clientWaitGroup.Done()
}

func (s *Server) shutdown() {
	go func() {
		for client := range s.Clients {
			client.Close()
		}
	}()

	for len(s.Clients) > 0 {
		select {
		case client := <-s.disconnect:
			s.removeClient(client)
		case conn := <-s.newConn:
			conn.Close()
		case <-s.Events.Receive():
		}
	}
}

func (s *Server) handleCommands(ctx context.Context) {
	pingTicker := time.NewTicker(1 * time.Minute)
	defer pingTicker.Stop()

	for {
		select {
		case msg := <-s.Events.Receive():
			// TODO: add proper client logging
			log.Printf("Received message from client: %s", msg.Raw)
			cmdCtx, cancel := context.WithTimeout(ctx, 1*time.Minute)
			go func() {
				defer cancel()
				s.commandRouter(msg, cmdCtx)
			}()
		case conn := <-s.newConn:
			log.Printf("New connection from %v", conn.RemoteAddr())
			s.clientWaitGroup.Add(1)
			s.handleConnection(conn)
		case client := <-s.disconnect:
			log.Printf("Disconnecting client: %v", client.Username)
			user, ok := s.Users.Get(client.Username)
			if ok && user.ActiveChannel != nil {
				delete(user.ActiveChannel.Members, user.Username)
				// TODO: Maybe specify that the user left because they got disconnected
				AnnounceLeftChannel(s, user.ActiveChannel, user.Username)
				sendStatusToBuddies(s, user, events.BuddyStatusOffline)
				user.ActiveChannel = nil
			}
			s.removeClient(client)
		case <-pingTicker.C:
			log.Println("Sending ping messages to all clients")
			for client := range s.Clients {
				go client.Ping()
			}
		case <-ctx.Done():
			log.Println("Shutting down server...")
			s.shutdown()
			return
		}
	}
}

func (s *Server) startListener(ctx context.Context) {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatal(err)
	}
	s.listener = listener
	go func() {
		<-ctx.Done()
		s.listener.Close()
	}()
}

func (s *Server) Run(ctx context.Context) {
	s.startListener(ctx)
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		defer wg.Done()
		s.handleConnections()
	}()

	s.handleCommands(ctx)
	wg.Wait()
	log.Println("Waiting for all clients to shut down cleanly")
	s.clientWaitGroup.Wait()
}
