package events

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
)

type Client struct {
	Username   string
	conn       net.Conn
	Outbound   chan Message
	disconnect chan<- *Client
	closeOnce  sync.Once
}

func NewClient(conn net.Conn, disconnect chan<- *Client) *Client {
	return &Client{
		conn:       conn,
		Outbound:   make(chan Message, 10), // buffered to prevent blocking
		disconnect: disconnect,
	}
}

func (c *Client) Close() error {
	var err error
	c.closeOnce.Do(func() {
		err = c.conn.Close()
		c.disconnect <- c
	})
	return err
}

func (c *Client) Ping() {
	select {
	case c.Outbound <- PingMessage:
	default:
		c.Close()
	}
}

func (c *Client) Run(events Engine) {
	go c.writeLoop()
	go c.readLoop(events)
}

func (c *Client) writeLoop() {
	defer c.Close()

	for msg := range c.Outbound {
		log.Printf("Sending message: %s", msg.Raw)
		_, err := fmt.Fprint(c.conn, msg.Raw+"\r\n")
		if err != nil {
			log.Printf("Error writing to client: %v", err)
			break
		}
	}
}

func (c *Client) readLoop(events Engine) {
	defer c.Close()

	scanner := bufio.NewScanner(c.conn)
	for scanner.Scan() {
		text := scanner.Text()
		msg := NewMessageFromClient(text, c)
		if msg.Command == QuitCommand {
			break
		}
		err := events.Send(msg)
		if err != nil {
			log.Printf("Error sending to server: %v", err)
			c.Outbound <- ServerErrMessage("ERROR: Server Busy.")
		}
	}
}
