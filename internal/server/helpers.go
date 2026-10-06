package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"maxim/internal/events"
	"maxim/internal/models"
	"net/url"
	"strings"
	"time"
)

func extractUserFromContext(ctx context.Context) *models.User {
	user, ok := ctx.Value("user").(*models.User)
	if !ok {
		return &models.User{}
	}
	return user
}

func sendStatusToBuddies(server *Server, currentUser *models.User, status events.BuddyStatus) {
	for client := range server.Clients {
		user, ok := server.Users.Get(client.Username)
		if !ok {
			continue
		}
		if _, ok = user.Buddies[currentUser.Username]; ok {
			client.Send(events.BuddyStatusMessage(currentUser.Username, status))
		}
	}
}

func processUserDetails(username string, server *Server, args []string) error {
	password := args[0]
	hashSalt, err := server.Hash.GenerateHash([]byte(password), []byte(server.Salt))
	if err != nil {
		log.Printf("Error generating hash for user %s: %v", username, err)
		return errors.New("error while processing password")
	}

	dob, err := time.Parse("02/01/06", args[4])
	if err != nil {
		log.Printf("Error parsing dob for user %s: %v", username, err)
		return errors.New("invalid DOB")
	}
	server.Users.Add(username, &models.User{
		Username:  username,
		Hash:      hashSalt.Hash,
		Salt:      hashSalt.Salt,
		FullName:  DecodeUserDetailField(args[1]),
		Gender:    args[2],
		Location:  DecodeUserDetailField(args[3]),
		DOB:       dob,
		Email:     DecodeUserDetailField(args[4]),
		Profile:   DecodeUserDetailField(args[5]),
		Signature: DecodeUserDetailField(args[6]),
		Buddies:   make(map[string]bool),
		Ignored:   make(map[string]bool),
	})
	return nil
}

func BroadcastToChannel(sender string, server *Server, channel *models.Channel, msg events.Message) {
	// FIXME: this is highly inefficient
	//  maybe keep a map of username->conn in server?
	for client := range server.Clients {
		if _, ok := channel.Members[client.Username]; ok {
			user, ok := server.Users.Get(client.Username)
			if _, ignored := user.Ignored[sender]; !ok || ignored {
				continue
			}
			client.Send(msg)
		}
	}
}

func AnnounceLeftChannel(server *Server, channel *models.Channel, username string) {
	BroadcastToChannel(username, server, channel, events.UserLeaveMessage(username))
	BroadcastToChannel(username, server, channel, events.ServerMessage(fmt.Sprintf("%s left channel", username)))
}

func AnnounceJoinedChannel(server *Server, channel *models.Channel, username string) {
	BroadcastToChannel(username, server, channel, events.UserJoinMessage(username))
	BroadcastToChannel(username, server, channel, events.ServerMessage(fmt.Sprintf("%s joined channel", username)))
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

func FormatWhoisAnswer(user *models.User) events.Message {
	now := time.Now()
	age := now.Year() - user.DOB.Year()
	if now.Month() < user.DOB.Month() || (now.Month() == user.DOB.Month() && now.Day() < user.DOB.Day()) {
		age--
	}
	return events.ServerMessage(fmt.Sprintf("[@%s] %s %s %dyo",
		user.Username,
		user.Gender,
		user.Location,
		age,
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
