package events

import (
	"bufio"
	"context"
	"errors"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

const MaxLineSize = 4096

type Client struct {
	conn       net.Conn
	outbound   chan Message
	disconnect chan<- *Client
	done       chan struct{}
	finished   chan struct{}
	closeOnce  sync.Once
	runOnce    sync.Once
}

func NewClient(conn net.Conn, disconnect chan<- *Client) *Client {
	return &Client{
		conn:       conn,
		disconnect: disconnect,
		done:       make(chan struct{}),
		finished:   make(chan struct{}),
		outbound:   make(chan Message, 256),
	}
}

func (c *Client) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *Client) Done() <-chan struct{} {
	return c.done
}

func (c *Client) Wait() {
	<-c.finished
}

func (c *Client) IsClosed() bool {
	select {
	case <-c.done:
		return true
	default:
		return false
	}
}

func (c *Client) Close() error {
	var err error
	c.closeOnce.Do(func() {
		close(c.done)
		err = c.conn.Close()
	})
	return err
}

func (c *Client) Send(msg Message) {
	if c.IsClosed() {
		return
	}
	select {
	case <-c.done:
	case c.outbound <- msg:
	default:
		_ = c.Close()
	}
}

func (c *Client) Ping() {
	c.Send(PingMessage())
}

func (c *Client) run(ctx context.Context, commands Engine) {
	defer close(c.finished)
	defer func() {
		select {
		case c.disconnect <- c:
		case <-ctx.Done():
		}
	}()

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error {
		select {
		case <-gCtx.Done():
			return c.Close()
		case <-c.done:
		}
		return nil
	})

	g.Go(func() error {
		return c.writeLoop(gCtx)
	})
	g.Go(func() error {
		return c.readLoop(gCtx, commands)
	})

	_ = g.Wait()
}

func (c *Client) Run(ctx context.Context, commands Engine) {
	c.runOnce.Do(func() {
		go c.run(ctx, commands)
	})
}

func (c *Client) writeLoop(ctx context.Context) error {
	defer func() {
		_ = c.Close()
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-c.done:
			return nil
		case msg, ok := <-c.outbound:
			if !ok {
				return nil
			}
			err := c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err != nil {
				log.Printf("Error setting write deadline: %v", err)
				return err
			}
			_, err = io.WriteString(c.conn, msg.Raw+"\r\n")
			if err != nil {
				log.Printf("Error writing to client: %v", err)
				return err
			}
		}
	}
}

func (c *Client) readLoop(ctx context.Context, commands Engine) error {
	defer func() {
		_ = c.Close()
	}()

	reader := bufio.NewReaderSize(c.conn, MaxLineSize+2)
	for {
		err := c.conn.SetReadDeadline(time.Now().Add(3 * time.Minute))
		if err != nil {
			log.Printf("Error setting read deadline: %v", err)
			return err
		}

		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if !c.IsClosed() && ctx.Err() == nil && !errors.Is(err, io.EOF) {
				log.Printf("Client read failed: %v", err)
			}
			return err
		}
		if isPrefix || len(line) > MaxLineSize {
			for isPrefix {
				_, isPrefix, err = reader.ReadLine()
				if err != nil {
					log.Printf("Client read failed: %v", err)
					return err
				}
			}
			c.Send(ServerErrMessage("Message exceeds maximum size."))
		}
		msg := NewMessageFromClient(string(line), c)
		if msg.Command == "" {
			continue
		}
		err = commands.Send(msg)
		if err != nil {
			if msg.Command == QuitCommand {
				return err
			}
			c.Send(ServerErrMessage("Server busy."))
		}
	}
}
