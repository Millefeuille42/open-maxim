package server

import (
	"fmt"
	"log"
	"maxim/internal/events"
	"maxim/internal/events/local"
	"maxim/internal/models"
	"maxim/internal/state"
	"maxim/internal/state/memory"
	"maxim/internal/utils"
	"net"
	"time"
)

type Server struct {
	Port int
	Hash *utils.Argon2idHash
	Salt string

	Clients map[*events.Client]bool

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
			log.Println(err)
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

func (s *Server) commandRouter(msg events.Message) {
	if handler, ok := handlers[msg.Command]; ok {
		handler(s, msg)
		return
	}
	log.Printf("Unknown command: %s", msg.Command)
}

func (s *Server) handleCommands() {
	pingTicker := time.NewTicker(1 * time.Minute)
	defer pingTicker.Stop()
	for {
		select {
		case msg := <-s.Events.Receive():
			// TODO: add proper client logging
			log.Printf("Received message from client: %s", msg.Raw)
			go s.commandRouter(msg)
		case conn := <-s.newConn:
			log.Printf("New connection from %v", conn.RemoteAddr())
			s.handleConnection(conn)
		case client := <-s.disconnect:
			log.Printf("Disconnecting client: %v", client.Username)
			if _, ok := s.Users.Get(client.Username); !ok {
				s.Users.Remove(client.Username)
			}
			delete(s.Clients, client)
		case <-pingTicker.C:
			log.Println("Sending ping messages to all clients")
			for client := range s.Clients {
				go client.Ping()
			}
		}
	}
}

func (s *Server) Run() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		log.Fatal(err)
	}
	s.listener = listener
	defer s.listener.Close()

	go s.handleConnections()
	s.handleCommands()
}
