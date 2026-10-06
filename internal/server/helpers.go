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

func validateUsername(name string, max int) bool {
	if len(name) <= 0 || len(name) > max {
		return false
	}

	return strings.IndexFunc(name, func(r rune) bool {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '_', r == '-', r == '.':
			return false
		default:
			return true
		}
	}) == -1
}

func parseUserDetails(args []string) (models.UserDetails, error) {
	details := models.UserDetails{}

	if len(args) < 8 {
		return details, errors.New("wrong number of profile arguments")
	}
	if args[0] == "" || len(args[0]) > 256 {
		return details, errors.New("invalid password length")
	}
	if len(args) == 9 && args[8] != "y" && args[8] != "n" {
		return details, errors.New("invalid avatar preference")
	}

	fields := make([]string, 8)
	for i, field := range args[1:8] {
		value, err := url.QueryUnescape(field)
		if err != nil {
			return details, errors.New("invalid profile encoding")
		}
		fields[i] = value
	}

	dob, err := time.Parse("02/01/06", fields[4])
	if err != nil {
		return details, errors.New("invalid DOB")
	}

	details = models.UserDetails{
		FullName:  fields[1],
		Gender:    fields[2],
		Location:  fields[3],
		DOB:       dob,
		Email:     fields[5],
		Profile:   fields[6],
		Signature: fields[7],
	}
	return details, nil
}

func processUserDetails(username string, server *Server, args []string) error {
	details, err := parseUserDetails(args)
	if err != nil {
		return err
	}
	hash, err := server.Hash.GenerateHash([]byte(args[0]))
	if err != nil {
		return errors.New("could not process password")
	}

	user, exists := server.Users.Get(username)
	if !exists {
		user = &models.User{Username: username, Buddies: make(map[string]bool), Ignored: make(map[string]bool)}
	}
	user.Details = details
	user.Password = *hash
	server.Users.Add(username, user)
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
	age := now.Year() - user.Details.DOB.Year()
	if now.Month() < user.Details.DOB.Month() ||
		(now.Month() == user.Details.DOB.Month() && now.Day() < user.Details.DOB.Day()) {
		age--
	}
	return events.ServerMessage(fmt.Sprintf("[@%s] %s %s %dyo",
		user.Username,
		user.Details.Gender,
		user.Details.Location,
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
