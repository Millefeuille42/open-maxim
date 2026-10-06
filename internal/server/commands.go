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
CHANNEL_REMOVED -> Not sure when it should happen

// Forum
FORUMS
GET_TOPIC_LIST
GET_POSTS_LIST
GET_FORUM_LIST
GETAVATARS
SMILEPACKAGE
ADDTOPIC
ADDPOST
AVATARIMAGE

// Advert
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
	events.BuddyCommand:         needsLoggedIn(buddy),
	events.IgnoreCommand:        needsLoggedIn(ignore),
	events.ReportCommand:        needsLoggedIn(report),

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
	user, ok := server.Users.Get(username)
	if ok {
		err := server.Hash.Compare(user.Password, []byte(password))
		if err == nil {
			msg.Origin.Username = username
			msg.Origin.Send(events.LoginOkMessage())
			sendStatusToBuddies(server, user, events.BuddyStatusOnline)
			return
		}
		log.Printf("Error generating hash for user %s: %v", username, err)
	}

	select {
	case <-time.After(time.Until(end)):
		msg.Origin.Send(events.LoginFailedMessage("Wrong username or password."))
	case <-ctx.Done():
		break
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
	if !validateUsername(username, 32) {
		msg.Origin.Send(events.RegisterFailedMessage("Invalid username."))
		return
	}
	if _, ok := server.Users.Get(username); ok {
		msg.Origin.Send(events.RegisterFailedMessage("Another user with this username already exists."))
		return
	}

	err := processUserDetails(username, server, msg.Args[1:])
	if err != nil {
		msg.Origin.Send(events.RegisterFailedMessage(err.Error()))
		return
	}
	msg.Origin.Username = username
	msg.Origin.Send(events.RegisterOkMessage())
	return
}

func ping(_ *Server, msg events.Message, _ context.Context) {
	msg.Origin.Send(events.PongMessage())
}

func say(server *Server, msg events.Message, ctx context.Context) {
	user := extractUserFromContext(ctx)
	BroadcastToChannel(msg.Origin.Username, server, user.ActiveChannel, FormatChannelMessage(
		user.ActiveChannel.Name,
		user.Username,
		strings.Join(msg.Args, " "),
	))
}

func chat(server *Server, msg events.Message, ctx context.Context) {
	channels := server.Channels.GetAll()
	chanNames := make([]string, 0, len(channels))
	for _, channel := range channels {
		chanNames = append(chanNames, channel.Name)
	}
	msg.Origin.Send(events.ReportChannelsMessage(chanNames))

	user := extractUserFromContext(ctx)
	buddies := make([]string, 0, len(user.Buddies))
	for buddy := range user.Buddies {
		buddies = append(buddies, buddy)
	}
	msg.Origin.Send(events.BuddyListMessage(buddies))

	ignoredUsers := make([]string, 0, len(user.Ignored))
	for ignored := range user.Ignored {
		ignoredUsers = append(ignoredUsers, ignored)
	}
	msg.Origin.Send(events.IgnoreListMessage(ignoredUsers))
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
	if _, ignored := user.Ignored[target.Username]; ignored {
		msg.Origin.Send(events.ServerErrMessage("User is ignored."))
		return
	}

	msg.Origin.Send(FormatWhisperSender(target.Username, message))

	if _, ignored := target.Ignored[user.Username]; ignored {
		return
	}
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

func newDetails(server *Server, msg events.Message, ctx context.Context) {
	if len(msg.Args) != 9 {
		msg.Origin.Send(events.ServerErrMessage("Wrong number of arguments."))
		return
	}
	user := extractUserFromContext(ctx)
	// TODO: Implement avatar management
	err := processUserDetails(user.Username, server, msg.Args)
	if err != nil {
		msg.Origin.Send(events.ServerErrMessage(err.Error()))
		return
	}
	msg.Origin.Send(events.UserDetailsMessage(user.Username, msg.Args))
	msg.Origin.Send(events.NewDetailsOkMessage())
}

func buddy(server *Server, msg events.Message, ctx context.Context) {
	if len(msg.Args) < 2 {
		msg.Origin.Send(events.ServerErrMessage("Wrong number of arguments."))
		return
	}

	if _, ok := server.Users.Get(msg.Args[1]); !ok {
		msg.Origin.Send(events.ServerErrMessage("User not found."))
		return
	}

	user := extractUserFromContext(ctx)
	target := msg.Args[1]

	switch strings.ToUpper(msg.Args[0]) {
	case "ADD":
		if _, ok := user.Buddies[target]; ok {
			msg.Origin.Send(events.ServerErrMessage("User is already your buddy."))
			return
		}
		user.Buddies[target] = true
		msg.Origin.Send(events.BuddyAddMessage(target))
		msg.Origin.Send(events.ServerMessage(fmt.Sprintf("%s is now your buddy!", target)))
	case "REMOVE":
		if _, ok := user.Buddies[target]; !ok {
			msg.Origin.Send(events.ServerErrMessage("User is not your buddy."))
			return
		}
		delete(user.Buddies, target)
		msg.Origin.Send(events.BuddyRemoveMessage(target))
		msg.Origin.Send(events.ServerMessage(fmt.Sprintf("%s is not your buddy anymore :(", target)))
	default:
		msg.Origin.Send(events.ServerErrMessage("Unsupported operation."))
	}
}

func ignore(server *Server, msg events.Message, ctx context.Context) {
	if len(msg.Args) < 2 {
		msg.Origin.Send(events.ServerErrMessage("Wrong number of arguments."))
		return
	}

	if _, ok := server.Users.Get(msg.Args[1]); !ok {
		msg.Origin.Send(events.ServerErrMessage("User not found."))
		return
	}

	user := extractUserFromContext(ctx)
	target := msg.Args[1]

	switch strings.ToUpper(msg.Args[0]) {
	case "ADD":
		if _, ok := user.Ignored[target]; ok {
			msg.Origin.Send(events.ServerErrMessage("User is already ignored."))
			return
		}
		user.Ignored[target] = true
		msg.Origin.Send(events.IgnoreAddMessage(target))
		msg.Origin.Send(events.ServerMessage(fmt.Sprintf("%s is now ignored.", target)))
	case "REMOVE":
		if _, ok := user.Ignored[target]; !ok {
			msg.Origin.Send(events.ServerErrMessage("User is not ignored."))
			return
		}
		delete(user.Ignored, target)
		msg.Origin.Send(events.IgnoreRemoveMessage(target))
		msg.Origin.Send(events.ServerMessage(fmt.Sprintf("%s is not ignored anymore.", target)))
	default:
		msg.Origin.Send(events.ServerErrMessage("Unsupported operation."))
	}
}

func report(_ *Server, msg events.Message, _ context.Context) {
	if len(msg.Args) < 1 {
		msg.Origin.Send(events.ServerErrMessage("Wrong number of arguments."))
		return
	}

	log.Printf("REPORT: User `%s` reported `%s`\n", msg.Origin.Username, msg.Args[0])
	msg.Origin.Send(events.ServerMessage(fmt.Sprintf("%s successfully reported.", msg.Args[0])))
}
