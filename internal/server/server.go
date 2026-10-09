package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maxim/internal/events"
	"maxim/internal/events/local"
	"maxim/internal/state"
	"maxim/internal/state/memory"
	"maxim/internal/utils"
	"net"
	"sync"
	"time"
)

const (
	maxClients         = 128
	maxPendingCommands = 32
	maxAuthWorkers     = 4
	maxAuthPerMinute   = 30
	maxRateEntries     = 4096
)

type session struct {
	client      *events.Client
	username    string
	channelName string
	pending     bool
	queue       []events.ClientMessage
}

type Server struct {
	Port     int
	Hash     *utils.Argon2idHash
	Users    state.UserEngine
	Channels state.ChannelEngine

	clients      map[*events.Client]*session
	online       map[string]*session
	commands     events.Engine
	newConn      chan net.Conn
	acceptErrors chan error
	disconnect   chan *events.Client
	authResults  chan authResult
	authSlots    chan struct{}
	authRates    map[string]authRate
	wg           sync.WaitGroup
}

func NewServer(port int) *Server {
	return &Server{
		Port:         port,
		Hash:         utils.NewArgon2idHash(1, 32, 64*1024, 32, 256),
		Users:        memory.NewUserStore(),
		Channels:     memory.NewChannelStore(),
		clients:      make(map[*events.Client]*session),
		online:       make(map[string]*session),
		commands:     local.NewLocalEngine(100),
		newConn:      make(chan net.Conn),
		acceptErrors: make(chan error, 1),
		disconnect:   make(chan *events.Client),
		authResults:  make(chan authResult, maxAuthWorkers),
		authSlots:    make(chan struct{}, maxAuthWorkers),
		authRates:    make(map[string]authRate),
	}
}

func (s *Server) Run(ctx context.Context) error {
	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return err
	}
	defer func() {
		_ = listener.Close()
	}()
	return s.Serve(ctx, listener)
}

func (s *Server) Serve(ctx context.Context, listener net.Listener) error {
	ctx, cancel := context.WithCancel(ctx)
	s.wg.Go(func() {
		s.handleConnections(ctx, listener)
	})

	defer func() {
		cancel()
		_ = listener.Close()
		for client := range s.clients {
			_ = client.Close()
		}
		for client := range s.clients {
			client.Wait()
		}
		s.wg.Wait()
		clear(s.clients)
		clear(s.online)
	}()
	return s.handleCommands(ctx)
}

func (s *Server) handleConnections(ctx context.Context, listener net.Listener) {
	log.Printf("Server listening on %s", listener.Addr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if errors.Is(err, net.ErrClosed) {
				select {
				case s.acceptErrors <- err:
				case <-ctx.Done():
				}
				return
			}
			log.Printf("Accept failed: %v", err)
			select {
			case <-ctx.Done():
				return
			case <-time.After(50 * time.Millisecond):
			}
			continue
		}
		select {
		case s.newConn <- conn:
		case <-ctx.Done():
			_ = conn.Close()
			return
		}
	}
}

func (s *Server) handleConnection(ctx context.Context, conn net.Conn) {
	if len(s.clients) >= maxClients {
		_ = conn.Close()
		return
	}
	client := events.NewClient(conn, s.disconnect)
	s.clients[client] = &session{
		client: client,
	}
	client.Run(ctx, s.commands)
}

func (s *Server) commandRouter(msg events.ClientMessage, ctx context.Context) {
	sess, ok := s.clients[msg.Origin]
	if !ok || msg.Origin.IsClosed() || ctx.Err() != nil {
		return
	}
	if sess.pending {
		if len(sess.queue) >= maxPendingCommands {
			_ = sess.client.Close()
			return
		}
		sess.queue = append(sess.queue, msg)
		return
	}
	log.Printf("Command %q from %s", msg.Command, sess.client.RemoteAddr())
	if handler, ok := handlers[msg.Command]; ok {
		handler(s, msg, newCommandContext(ctx, sess))
		return
	}
	msg.Origin.Send(events.ServerErrMessage("Unknown command."))
}

func (s *Server) removeClient(client *events.Client) {
	sess, ok := s.clients[client]
	if !ok {
		return
	}
	delete(s.clients, client)
	if sess.username == "" {
		client.Wait()
		return
	}
	delete(s.online, sess.username)
	if sess.channelName != "" {
		// TODO: Maybe specify that the user left because they got disconnected
		channel, ok := s.Channels.Get(sess.channelName)
		if !ok {
			log.Println("Could not find channel", sess.channelName)
			sess.client.Send(events.ServerErrMessage("Failed to send message."))
			return
		}
		AnnounceLeftChannel(s, channel, sess.username)
		delete(channel.Members, sess.username)
		_ = s.Channels.Update(channel)
	}
	if user, ok := s.Users.Get(sess.username); ok {
		sendStatusToBuddies(s, user, events.BuddyStatusOffline)
	}
	client.Wait()
}

func (s *Server) handleCommands(ctx context.Context) error {
	pingTicker := time.NewTicker(1 * time.Minute)
	defer pingTicker.Stop()

	for {
		if ctx.Err() != nil {
			return nil
		}
		select {
		case <-ctx.Done():
			return nil
		case err := <-s.acceptErrors:
			return fmt.Errorf("accept connection: %w", err)
		case msg, ok := <-s.commands.Receive():
			if !ok {
				return errors.New("command queue closed")
			}
			s.commandRouter(msg, ctx)
		case result := <-s.authResults:
			s.finishAuth(result, ctx)
		case conn := <-s.newConn:
			if ctx.Err() != nil {
				_ = conn.Close()
				continue
			}
			log.Printf("New connection from %v", conn.RemoteAddr())
			s.handleConnection(ctx, conn)
		case client := <-s.disconnect:
			log.Printf("Disconnecting client: %v", client.RemoteAddr())
			s.removeClient(client)
		case now := <-pingTicker.C:
			log.Println("Sending ping messages to all clients")
			for client := range s.clients {
				client.Ping()
			}
			for ip, rate := range s.authRates {
				if now.Sub(rate.start) >= time.Minute {
					delete(s.authRates, ip)
				}
			}
		}
	}
}
