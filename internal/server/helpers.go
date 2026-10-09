package server

import (
	"context"
	"errors"
	"fmt"
	"maxim/internal/events"
	"maxim/internal/models"
	"net/url"
	"strings"
	"time"
)

type commandContext struct {
	context.Context
	session *session
	user    *models.User
}

func newCommandContext(ctx context.Context, sess *session) *commandContext {
	return &commandContext{
		Context: ctx,
		session: sess,
	}
}

func validateName(name string, max int) bool {
	if len(name) == 0 || len(name) > max {
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

	dob, err := time.Parse("02/01/06", fields[3])
	if err != nil {
		return details, errors.New("invalid DOB")
	}

	details = models.UserDetails{
		FullName:  fields[0],
		Gender:    fields[1],
		Location:  fields[2],
		DOB:       dob,
		Email:     fields[4],
		Profile:   fields[5],
		Signature: fields[6],
	}
	return details, nil
}

func broadcastToChannel(server *Server, channel *models.Channel, sender string, msg events.Message) {
	for username := range channel.Members {
		sess, ok := server.online[username]
		if !ok || sess.channelName != channel.Name {
			continue
		}
		if sender != "" && sender != username {
			user, ok := server.Users.Get(username)
			if !ok || user.Ignored[sender] {
				continue
			}
		}
		sess.client.Send(msg)
	}
}

func sendStatusToBuddies(server *Server, currentUser *models.User, status events.BuddyStatus) {
	for username, sess := range server.online {
		user, ok := server.Users.Get(username)
		if ok && user.Buddies[currentUser.Username] {
			sess.client.Send(events.BuddyStatusMessage(currentUser.Username, status))
		}
	}
}

func AnnounceLeftChannel(server *Server, channel *models.Channel, username string) {
	broadcastToChannel(server, channel, "", events.UserLeaveMessage(username))
	broadcastToChannel(server, channel, "", events.ServerMessage(fmt.Sprintf("%s left the channel", username)))
}

func AnnounceJoinedChannel(server *Server, channel *models.Channel, username string) {
	broadcastToChannel(server, channel, "", events.UserJoinMessage(username))
	broadcastToChannel(server, channel, "", events.ServerMessage(fmt.Sprintf("%s joined the channel", username)))
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

func buddyStatus(server *Server, username string) events.BuddyStatus {
	sess, ok := server.online[username]
	if ok && !sess.client.IsClosed() {
		return events.BuddyStatusOnline
	}
	return events.BuddyStatusOffline
}

func saveUser(server *Server, msg events.ClientMessage, user *models.User) bool {
	if err := server.Users.Update(user); err != nil {
		msg.Origin.Send(events.ServerErrMessage("Could not update account."))
		return false
	}
	return true
}

func needsLoggedIn(handler commandHandler) commandHandler {
	return func(server *Server, msg events.ClientMessage, ctx *commandContext) {
		if ctx.session.username == "" {
			msg.Origin.Send(events.ServerErrMessage("You need to log in first."))
			return
		}
		user, ok := server.Users.Get(ctx.session.username)
		if !ok {
			msg.Origin.Send(events.ServerErrMessage("You are logged in with an invalid user."))
			return
		}
		ctx.user = user
		handler(server, msg, ctx)
	}
}

func needsChannel(handler commandHandler) commandHandler {
	return needsLoggedIn(func(server *Server, msg events.ClientMessage, ctx *commandContext) {
		if ctx.session.channelName == "" {
			msg.Origin.Send(events.ServerErrMessage("You need to be in a channel to send a message."))
			return
		}
		handler(server, msg, ctx)
	})
}

func noop(_ *Server, _ events.ClientMessage, _ *commandContext) {}
