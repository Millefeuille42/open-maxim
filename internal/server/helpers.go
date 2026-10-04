package server

import (
	"context"
	"fmt"
	"log"
	"maxim/internal/events"
	"maxim/internal/models"
)

func extractUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value("user").(*models.User)
	if !ok {
		return &models.User{}
	}
	return user
}

func BroadcastToChannel(server *Server, channel *models.Channel, msg events.Message) {
	// FIXME: this is highly inefficient
	//  maybe keep a map of username->conn in server?
	for client := range server.Clients {
		if _, ok := channel.Members[client.Username]; ok {
			client.Send(msg)
		}
	}
}

func AnnounceLeftChannel(server *Server, channel *models.Channel, username string) {
	BroadcastToChannel(server, channel, events.UserLeaveMessage(username))
	BroadcastToChannel(server, channel, events.ServerMessage(fmt.Sprintf("%s left channel", username)))
}

func AnnounceJoinedChannel(server *Server, channel *models.Channel, username string) {
	BroadcastToChannel(server, channel, events.UserJoinMessage(username))
	BroadcastToChannel(server, channel, events.ServerMessage(fmt.Sprintf("%s joined channel", username)))
}

func FormatChannelMessage(channel, username, message string) events.Message {
	return events.UserMessage(fmt.Sprintf("[%s] @%s: %s",
		channel,
		username,
		message,
	))
}

func FormatWhisperSender(target, message string) events.Message {
	return events.UserMessage(fmt.Sprintf("[To @%s] %s",
		target,
		message,
	))
}

func FormatWhisperTarget(sender, message string) events.Message {
	return events.UserMessage(fmt.Sprintf("[@%s] %s",
		sender,
		message,
	))
}

func needsLoggedIn(handler commandHandler) commandHandler {
	// TODO: check how to handle if the user is logged in
	//  another server of the pool, should not be an issue
	//  if the server that spreads the message includes
	//  user info in the pub/sub message
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
			return
		}

		handler(server, msg, context.WithValue(ctx, "channel", user.ActiveChannel))
	})
}

func noop(_ *Server, _ events.Message, _ context.Context) {}
