package server

import (
	"context"
	"fmt"
	"log"
	"maxim/internal/events"
	"maxim/internal/models"
	"strings"
	"time"
)

/* TODO implement:
// Chat
WHOIS -> Will likely send a SERVER_MSG containing the info
CHANNEL_REMOVED -> Not sure when it should happen

NEWDETAILS -> gets sent when submitting new details
USER_DETAILS -> tell client to apply changes to the UI
NEWDETAILS_OK -> commit changes to memory card

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


// Generic
CLIENT_PROTOCOL -> skipped
CLIENT_TYPE -> skipped
PONG -> skipped

// Forum
FORUMS
GET_TOPIC_LIST
GET_POSTS_LIST
GET_FORUM_LIST
GETAVATARS
SMILEPACKAGE
ADDTOPIC
ADDPOST

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
	events.DirectMessageCommand: needsLoggedIn(directMessage),
	events.JoinCommand:          needsLoggedIn(join),
	events.CreateChannelCommand: needsLoggedIn(create),
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

func say(server *Server, msg events.Message, ctx context.Context) {
	user := extractUserFromContext(ctx)
	BroadcastToChannel(server, user.ActiveChannel, FormatChannelMessage(
		user.ActiveChannel.Name,
		user.Username,
		strings.Join(msg.Args, " "),
	))

}

func chat(server *Server, msg events.Message, _ context.Context) {
	channels := server.Channels.GetAll()
	chanNames := make([]string, 0, len(channels))
	for _, channel := range channels {
		chanNames = append(chanNames, channel.Name)
	}

	msg.Origin.Send(events.ReportChannelsMessage(chanNames))
	// TODO: Send buddy list
	// TODO: Send ignore list
}

func join(server *Server, msg events.Message, ctx context.Context) {
	user := extractUserFromContext(ctx)
	if len(msg.Args) < 1 {
		// TODO: do not do if the originator doesn't live on the server?
		msg.Origin.Send(events.ServerErrMessage("Missing channel name."))
		return
	}
	channelName := msg.Args[0]
	channel, ok := server.Channels.Get(channelName)
	if !ok {
		msg.Origin.Send(events.ServerErrMessage("Channel not found."))
		return
	}
	if user.ActiveChannel != nil {
		delete(user.ActiveChannel.Members, user.Username)
		AnnounceLeftChannel(server, user.ActiveChannel, user.Username)
	}
	user.ActiveChannel = channel
	AnnounceJoinedChannel(server, user.ActiveChannel, user.Username)
	channel.Members[user.Username] = true
	users := make([]string, 0, len(channel.Members))
	for s := range channel.Members {
		users = append(users, s)
	}
	msg.Origin.Send(events.ReportUsersMessage(users))
	msg.Origin.Send(events.ServerMessage(fmt.Sprintf("Joined channel %s.", channelName)))
}

func create(server *Server, msg events.Message, _ context.Context) {
	if len(msg.Args) < 1 {
		msg.Origin.Send(events.ServerErrMessage("Missing channel name."))
		return
	}
	channelName := msg.Args[0]
	_, ok := server.Channels.Get(channelName)
	if ok {
		msg.Origin.Send(events.ServerErrMessage("Channel already exists."))
		return
	}

	if !strings.HasPrefix(channelName, "#") {
		msg.Origin.Send(events.ServerErrMessage("Channel name must start with '#'."))
		return
	}

	server.Channels.Add(channelName, models.NewChannel(channelName))
	msg.Origin.Send(events.ServerMessage(fmt.Sprintf("Created channel: %s", channelName)))
	msg.Origin.Send(events.ChannelAddedMessage(channelName))
}

func directMessage(server *Server, msg events.Message, ctx context.Context) {
	if len(msg.Args) < 2 {
		msg.Origin.Send(events.ServerErrMessage("Wrong number of arguments."))
		return
	}
	user := extractUserFromContext(ctx)
	userName := msg.Args[0]
	message := strings.Join(msg.Args[1:], " ")
	target, ok := server.Users.Get(userName)
	if !ok {
		msg.Origin.Send(events.ServerErrMessage("User not found."))
		return
	}
	msg.Origin.Send(FormatWhisperSender(target.Username, message))

	// FIXME: this is highly inefficient
	//  maybe keep a map of username->conn in server?
	for client := range server.Clients {
		if client.Username == target.Username {
			client.Send(FormatWhisperTarget(user.Username, message))
		}
	}
}

func whois(server *Server, msg events.Message, _ context.Context) {
	if len(msg.Args) < 1 {
		msg.Origin.Send(events.ServerErrMessage("Missing user name."))
		return
	}
	user, ok := server.Users.Get(msg.Args[0])
	if !ok {
		msg.Origin.Send(events.ServerErrMessage("User not found."))
		return
	}

	msg.Origin.Send(FormatWhoisAnswer(user))
}

func newDetails(server *Server, msg events.Message, _ context.Context) {
	//TODO implement me
	panic("implement me")
}
