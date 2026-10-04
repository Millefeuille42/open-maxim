package server

import (
	"log"
	"maxim/internal/events"
	"maxim/internal/models"
	"strings"
	"time"
)

type commandHandler func(*Server, events.Message)

var handlers = map[string]commandHandler{
	events.LoginCommand:    login,
	events.PingCommand:     ping,
	events.RegisterCommand: register,

	events.SayCommand: needsLoggedIn(say),

	events.PongCommand:           noop,
	events.ClientProtocolCommand: noop,
	events.ClientTypeCommand:     noop,
}

func login(server *Server, msg events.Message) {
	if len(msg.Args) != 2 {
		msg.Origin.Outbound <- events.LoginFailedMessage("Wrong number of arguments.")
		return
	}

	start := time.Now()
	end := start.Add(1 * time.Second)

	if msg.Origin.Username != "" {
		msg.Origin.Outbound <- events.LoginFailedMessage("You are already logged in.")
		return
	}

	username := msg.Args[0]
	password := msg.Args[1]

	if user, ok := server.Users.Get(username); ok {
		err := server.Hash.Compare(user.Hash, user.Salt, []byte(password))
		if err == nil {
			msg.Origin.Username = username
			msg.Origin.Outbound <- events.LoginOkMessage
			return
		}
		log.Printf("Error generating hash for user %s: %v", username, err)
	}

	time.Sleep(time.Until(end))
	msg.Origin.Outbound <- events.LoginFailedMessage("Wrong username or password.")
}

func register(server *Server, msg events.Message) {
	if len(msg.Args) != 9 {
		msg.Origin.Outbound <- events.RegisterFailedMessage("Wrong number of arguments.")
		return
	}

	if msg.Origin.Username != "" {
		msg.Origin.Outbound <- events.RegisterFailedMessage("You are already logged in.")
		return
	}

	username := msg.Args[0]
	if _, ok := server.Users.Get(username); ok {
		msg.Origin.Outbound <- events.RegisterFailedMessage("Another user with this username already exists.")
		return
	}

	password := msg.Args[1]
	hashSalt, err := server.Hash.GenerateHash([]byte(password), []byte(server.Salt))
	if err != nil {
		log.Printf("Error generating hash for user %s: %v", username, err)
		msg.Origin.Outbound <- events.RegisterFailedMessage("Error while registering user.")
		return
	}

	dob, err := time.Parse("02/01/06", msg.Args[5])
	if err != nil {
		msg.Origin.Outbound <- events.RegisterFailedMessage("Invalid DOB.")
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
	msg.Origin.Outbound <- events.RegisterOkMessage
	return
}

func ping(_ *Server, msg events.Message) {
	msg.Origin.Outbound <- events.PongMessage
}

func say(server *Server, msg events.Message) {
	for client := range server.Clients {
		go func() {
			client.Outbound <- events.UserMessage(msg.Origin.Username, strings.Join(msg.Args, " "))
		}()
	}
}
