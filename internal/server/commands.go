package server

import (
	"context"
	"log"
	"maxim/internal/events"
	"maxim/internal/models"
	"strings"
	"time"
)

/* TODO implement:
// Chat
CHAT
JOIN
MSG
CREATE
WHOIS
NEWDETAILS

CHANNEL_ADDED
CHANNEL_REMOVED
USER_JOIN
USER_LEAVE
SERVER_MSG

NEWDETAILS_OK
USER_DETAILS -> ???

REPORT
IGNORE ADD
IGNORE REMOVE
BUDDY ADD
BUDDY REMOVE
IGNORE_ADD
IGNORE_DEL
IGNORE_LIST
BUDDY_ADD
BUDDY_DEL
BUDDY_LIST
BUDDY_STATUS
REPORT_USERS
REPORT_CHANNELS


// Generic
CLIENT_PROTOCOL -> skipped
CLIENT_TYPE -> skipped
PONG -> skipped
ADDTOPIC
ADDPOST

// Forum
FORUMS
GET_TOPIC_LIST
GET_POSTS_LIST
GET_FORUM_LIST
GETAVATARS
SMILEPACKAGE

// Others / unidentified
PS2_SETTINGS
PASSWORD
CONFIRM_CHANGES
SETTINGS
AVATARIMAGE
ADVERT_VALIDITY
ADVERT_TIME
ADVERT_IMAGE
*/

type commandHandler func(*Server, events.Message, context.Context)

var handlers = map[events.Command]commandHandler{
	events.LoginCommand:    login,
	events.PingCommand:     ping,
	events.RegisterCommand: register,

	events.SayCommand:           needsChannel(say),
	events.ChatCommand:          needsLoggedIn(chat),
	events.JoinCommand:          needsLoggedIn(join),
	events.CreateChannelCommand: needsLoggedIn(create),
	events.DirectMessageCommand: needsLoggedIn(directMessage),
	events.WhoisCommand:         needsLoggedIn(whois),
	events.NewDetailsCommand:    needsLoggedIn(newDetails),

	events.PongCommand:           noop,
	events.ClientProtocolCommand: noop,
	events.ClientTypeCommand:     noop,
}

func login(server *Server, msg events.Message, ctx context.Context) {
	if len(msg.Args) != 2 {
		msg.Origin.Send(events.LoginFailedMessage("Wrong number of arguments."))
		return
	}

	start := time.Now()
	end := start.Add(1 * time.Second)

	if msg.Origin.Username != "" {
		msg.Origin.Send(events.LoginFailedMessage("You are already logged in."))
		return
	}

	username := msg.Args[0]
	password := msg.Args[1]

	if user, ok := server.Users.Get(username); ok {
		err := server.Hash.Compare(user.Hash, user.Salt, []byte(password))
		if err == nil {
			msg.Origin.Username = username
			msg.Origin.Send(events.LoginOkMessage())
			return
		}
		log.Printf("Error generating hash for user %s: %v", username, err)
	}

	select {
	case <-time.After(time.Until(end)):
		msg.Origin.Send(events.LoginFailedMessage("Wrong username or password."))
	case <-ctx.Done():
		return
	}
}

func register(server *Server, msg events.Message, _ context.Context) {
	if len(msg.Args) != 9 {
		msg.Origin.Send(events.RegisterFailedMessage("Wrong number of arguments."))
		return
	}

	if msg.Origin.Username != "" {
		msg.Origin.Send(events.RegisterFailedMessage("You are already logged in."))
		return
	}

	username := msg.Args[0]
	if _, ok := server.Users.Get(username); ok {
		msg.Origin.Send(events.RegisterFailedMessage("Another user with this username already exists."))
		return
	}

	password := msg.Args[1]
	hashSalt, err := server.Hash.GenerateHash([]byte(password), []byte(server.Salt))
	if err != nil {
		log.Printf("Error generating hash for user %s: %v", username, err)
		msg.Origin.Send(events.RegisterFailedMessage("Error while registering user."))
		return
	}

	dob, err := time.Parse("02/01/06", msg.Args[5])
	if err != nil {
		msg.Origin.Send(events.RegisterFailedMessage("Invalid DOB."))
		return
	}
	server.Users.Add(username, &models.User{
		Username:  username,
		Hash:      hashSalt.Hash,
		Salt:      hashSalt.Salt,
		FullName:  msg.Args[2],
		Gender:    msg.Args[3],
		Location:  msg.Args[4],
		DOB:       dob,
		Email:     msg.Args[6],
		Profile:   msg.Args[7],
		Signature: msg.Args[8],
	})
	msg.Origin.Username = username
	msg.Origin.Send(events.RegisterOkMessage())
	return
}

func ping(_ *Server, msg events.Message, _ context.Context) {
	msg.Origin.Send(events.PongMessage())
}

func say(server *Server, msg events.Message, _ context.Context) {
	for client := range server.Clients {
		client.Send(events.UserMessage(msg.Origin.Username, strings.Join(msg.Args, " ")))
	}
}

func chat(server *Server, msg events.Message, _ context.Context) {
	//TODO implement me
	panic("implement me")
}

func join(server *Server, msg events.Message, _ context.Context) {
	//TODO implement me
	panic("implement me")
}

func directMessage(server *Server, msg events.Message, _ context.Context) {
	//TODO implement me
	panic("implement me")
}

func create(server *Server, msg events.Message, _ context.Context) {
	//TODO implement me
	panic("implement me")
}

func whois(server *Server, msg events.Message, _ context.Context) {
	//TODO implement me
	panic("implement me")
}

func newDetails(server *Server, msg events.Message, _ context.Context) {
	//TODO implement me
	panic("implement me")
}
