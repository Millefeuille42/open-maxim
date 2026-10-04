package server

import (
	"context"
	"log"
	"maxim/internal/events"
	"maxim/internal/models"
)

func needsLoggedIn(handler commandHandler) commandHandler {
	return func(server *Server, msg events.Message, ctx context.Context) {
		if msg.Origin.Username == "" {
			log.Printf("User %s needs logging in", msg.Origin.Username)
			msg.Origin.Send(events.ServerErrMessage("You need to log in first."))
			return
		}
		user, ok := server.Users.Get(msg.Origin.Username)
		if !ok {
			msg.Origin.Send(events.ServerErrMessage("You are logged in with an invalid user."))
			return
		}

		handler(server, msg, context.WithValue(ctx, "user", user))
	}
}

func needsChannel(handler commandHandler) commandHandler {
	return needsLoggedIn(func(server *Server, msg events.Message, ctx context.Context) {
		user, ok := ctx.Value("user").(*models.User)
		if !ok {
			msg.Origin.Send(events.ServerErrMessage("You are logged in with an invalid user."))
			return
		}
		if user.ActiveChannel == nil {
			msg.Origin.Send(events.ServerErrMessage("You need to be in a channel to send a message."))
		}

		handler(server, msg, context.WithValue(ctx, "channel", user.ActiveChannel))
	})
}

func noop(_ *Server, _ events.Message, _ context.Context) {}
