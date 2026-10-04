package main

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
	"os"
	"strings"
	"time"
)

// TODO: add context, cancellation etc

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

type commandHandler func(*Server, events.Message)

var commands = map[string]commandHandler{
	events.LoginCommand:    login,
	events.PingCommand:     ping,
	events.RegisterCommand: register,

	events.SayCommand: needsLoggedIn(say),

	events.PongCommand:           noop,
	events.ClientProtocolCommand: noop,
	events.ClientTypeCommand:     noop,
}

func needsLoggedIn(handler commandHandler) commandHandler {
	return func(server *Server, msg events.Message) {
		if msg.Origin.Username == "" {
			log.Printf("User %s needs logging in", msg.Origin.Username)
			msg.Origin.Outbound <- events.ServerErrMessage("You need to log in first.")
			return
		}

		if _, ok := server.Users.Get(msg.Origin.Username); !ok {
			msg.Origin.Outbound <- events.ServerErrMessage("You are logged in with an invalid user.")
			return
		}
		handler(server, msg)
	}
}

func noop(_ *Server, _ events.Message) {}

func login(server *Server, msg events.Message) {
	if len(msg.Args) != 2 {
		msg.Origin.Outbound <- events.LoginFailedMessage("Wrong number of arguments.")
		return
	}

	start := time.Now()
	end := start.Add(1 * time.Second)

	if msg.Origin.Username != "" {
		msg.Origin.Outbound <- events.LoginFailedMessage("You are already logged in.")
		return
	}

	username := msg.Args[0]
	password := msg.Args[1]

	if user, ok := server.Users.Get(username); ok {
		err := server.Hash.Compare(user.Hash, user.Salt, []byte(password))
		if err == nil {
			msg.Origin.Username = username
			msg.Origin.Outbound <- events.LoginOkMessage
			return
		}
		log.Printf("Error generating hash for user %s: %v", username, err)
	}

	time.Sleep(time.Until(end))
	msg.Origin.Outbound <- events.LoginFailedMessage("Wrong username or password.")
}

func register(server *Server, msg events.Message) {
	if len(msg.Args) != 9 {
		msg.Origin.Outbound <- events.RegisterFailedMessage("Wrong number of arguments.")
		return
	}

	if msg.Origin.Username != "" {
		msg.Origin.Outbound <- events.RegisterFailedMessage("You are already logged in.")
		return
	}

	username := msg.Args[0]
	if _, ok := server.Users.Get(username); ok {
		msg.Origin.Outbound <- events.RegisterFailedMessage("Another user with this username already exists.")
		return
	}

	password := msg.Args[1]
	hashSalt, err := server.Hash.GenerateHash([]byte(password), []byte(server.Salt))
	if err != nil {
		log.Printf("Error generating hash for user %s: %v", username, err)
		msg.Origin.Outbound <- events.RegisterFailedMessage("Error while registering user.")
		return
	}

	dob, err := time.Parse("02/01/06", msg.Args[5])
	if err != nil {
		msg.Origin.Outbound <- events.RegisterFailedMessage("Invalid DOB.")
		return
	}
	server.Users.Add(username, &models.User{
		Username:  username,
		Hash:      hashSalt.Hash,
		Salt:      hashSalt.Salt,
		FullName:  msg.Args[2],
		Gender:    msg.Args[3],
		Location:  msg.Args[4],
		DOB:       dob,
		Email:     msg.Args[6],
		Profile:   msg.Args[7],
		Signature: msg.Args[8],
	})
	msg.Origin.Username = username
	msg.Origin.Outbound <- events.RegisterOkMessage
	return
}

func ping(_ *Server, msg events.Message) {
	msg.Origin.Outbound <- events.PongMessage
}

func say(server *Server, msg events.Message) {
	for client := range server.Clients {
		go func() {
			client.Outbound <- events.UserMessage(msg.Origin.Username, strings.Join(msg.Args, " "))
		}()
	}
}

func (s *Server) commandRouter(msg events.Message) {
	if handler, ok := commands[msg.Command]; ok {
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

func main() {
	// TODO add argument parsing?
	salt := os.Getenv("MAXIM_SALT")
	if len(salt) == 0 {
		log.Fatal("MAXIM_SALT environment variable not set")
	}
	server := NewServer(2002, salt)
	server.Run()
}
