package server

import (
	"log"
	"maxim/internal/events"
)

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
