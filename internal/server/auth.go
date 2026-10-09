package server

import (
	"bytes"
	"context"
	"maxim/internal/events"
	"maxim/internal/models"
	"maxim/internal/utils"
	"net"
	"time"
)

type authResult struct {
	client  *events.Client
	command events.Command
	user    *models.User
	err     error
	reply   events.Message
}

type authRate struct {
	attempts int
	start    time.Time
}

func (s *Server) startAuth(msg events.ClientMessage, ctx *commandContext, work func() (*models.User, error)) {
	ip := msg.Origin.RemoteAddr().String()
	if host, _, err := net.SplitHostPort(ip); err == nil {
		ip = host
	}

	now := time.Now()
	rate, exists := s.authRates[ip]
	if !exists && len(s.authRates) >= maxRateEntries {
		authError(msg.Origin, msg.Command, "Server busy.")
		return
	}

	if now.Sub(rate.start) >= time.Minute {
		rate = authRate{start: now}
	}

	if rate.attempts >= maxAuthPerMinute {
		authError(msg.Origin, msg.Command, "Too many authentication attempts. Try again later.")
		return
	}

	select {
	case s.authSlots <- struct{}{}:
	default:
		authError(msg.Origin, msg.Command, "Server busy.")
		return
	}

	rate.attempts++
	s.authRates[ip] = rate
	ctx.session.pending = true

	var reply events.Message
	if msg.Command == events.NewDetailsCommand {
		reply = events.UserDetailsMessage(ctx.session.username, msg.Args)
	}

	s.wg.Go(func() {
		defer func() {
			<-s.authSlots
		}()

		user, err := work()
		result := authResult{
			client:  msg.Origin,
			command: msg.Command,
			user:    user,
			err:     err,
			reply:   reply,
		}
		select {
		case <-ctx.Done():
		case <-msg.Origin.Done():
		case s.authResults <- result:
		}
	})
}

func authError(client *events.Client, command events.Command, reason string) {
	switch command {
	case events.LoginCommand:
		client.Send(events.LoginFailedMessage(reason))
	case events.RegisterCommand:
		client.Send(events.RegisterFailedMessage(reason))
	default:
		client.Send(events.ServerErrMessage(reason))
	}
}

func (s *Server) finishAuth(result authResult, ctx context.Context) {
	sess, exists := s.clients[result.client]
	if !exists || result.client.IsClosed() || ctx.Err() != nil {
		return
	}
	sess.pending = false

	if result.err != nil {
		authError(result.client, result.command, result.err.Error())
		return
	}

	switch result.command {
	case events.LoginCommand:
		current, exists := s.Users.Get(result.user.Username)
		if !exists || current.Password.Params != result.user.Password.Params ||
			!bytes.Equal(current.Password.Hash, result.user.Password.Hash) ||
			!bytes.Equal(current.Password.Salt, result.user.Password.Salt) {
			// a password change at the same time can result in an error
			authError(result.client, result.command, utils.ErrPasswordMismatch.Error())
			return
		}

		if _, exists := s.online[result.user.Username]; exists {
			authError(result.client, result.command, "User is already logged in.")
			return
		}

		s.authenticate(sess, result.user)
		result.client.Send(events.LoginOkMessage())
		sendStatusToBuddies(s, result.user, events.BuddyStatusOnline)
	case events.RegisterCommand:
		if err := s.Users.Create(result.user); err != nil {
			authError(result.client, result.command, err.Error())
			return
		}

		s.authenticate(sess, result.user)
		result.client.Send(events.RegisterOkMessage())
	case events.NewDetailsCommand:
		user, ok := s.Users.Get(sess.username)
		if !ok {
			authError(result.client, result.command, "User not found.")
			return
		}
		user.Details, user.Password = result.user.Details, result.user.Password

		if err := s.Users.Update(user); err != nil {
			authError(result.client, result.command, "Could not update profile.")
			return
		}

		result.client.Send(result.reply)
		result.client.Send(events.NewDetailsOkMessage())
	}

	for len(sess.queue) > 0 && !sess.pending && !result.client.IsClosed() && ctx.Err() == nil {
		msg := sess.queue[0]
		sess.queue[0] = events.ClientMessage{}
		sess.queue = sess.queue[1:]
		s.commandRouter(msg, ctx)
	}
	if len(sess.queue) == 0 {
		sess.queue = nil
	}
}

func (s *Server) authenticate(sess *session, user *models.User) {
	sess.username = user.Username
	s.online[user.Username] = sess
}
