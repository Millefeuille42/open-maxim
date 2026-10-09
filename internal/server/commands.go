package server

import (
	"errors"
	"fmt"
	"log"
	"maxim/internal/events"
	"maxim/internal/models"
	"maxim/internal/utils"
	"slices"
	"strings"
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

type commandHandler func(*Server, events.ClientMessage, *commandContext)

var handlers = map[events.Command]commandHandler{
	events.LoginCommand:          login,
	events.PingCommand:           ping,
	events.RegisterCommand:       register,
	events.QuitCommand:           quit,
	events.SayCommand:            needsChannel(say),
	events.ChatCommand:           needsLoggedIn(chat),
	events.DirectMessageCommand:  needsLoggedIn(directMessage),
	events.JoinCommand:           needsLoggedIn(join),
	events.CreateChannelCommand:  needsLoggedIn(create),
	events.WhoisCommand:          needsLoggedIn(whois),
	events.NewDetailsCommand:     needsLoggedIn(newDetails),
	events.BuddyCommand:          needsLoggedIn(buddy),
	events.IgnoreCommand:         needsLoggedIn(ignore),
	events.ReportCommand:         needsLoggedIn(report),
	events.PongCommand:           noop,
	events.ClientProtocolCommand: noop,
	events.ClientTypeCommand:     noop,
}

func login(server *Server, msg events.ClientMessage, ctx *commandContext) {
	if len(msg.Args) != 2 || len(msg.Args[1]) > 256 {
		authError(msg.Origin, msg.Command, "Wrong number of arguments or invalid password length.")
		return
	}

	if ctx.session.username != "" {
		authError(msg.Origin, msg.Command, "You are already logged in.")
		return
	}

	user, ok := server.Users.Get(msg.Args[0])
	password := []byte(msg.Args[1])
	server.startAuth(msg, ctx, func() (*models.User, error) {
		if !ok {
			// still compute hash to mitigate user listing attacks
			_, _ = server.Hash.GenerateHash(password)
			return nil, utils.ErrPasswordMismatch
		}
		if err := server.Hash.Compare(user.Password, password); err != nil {
			return nil, utils.ErrPasswordMismatch
		}
		return user, nil
	})
}

func register(server *Server, msg events.ClientMessage, ctx *commandContext) {
	if len(msg.Args) != 9 {
		msg.Origin.Send(events.RegisterFailedMessage("Wrong number of arguments."))
		return
	}

	if ctx.session.username != "" {
		msg.Origin.Send(events.RegisterFailedMessage("You are already logged in."))
		return
	}

	username := msg.Args[0]
	if !validateName(username, 32) {
		msg.Origin.Send(events.RegisterFailedMessage("Invalid username."))
		return
	}

	if _, ok := server.Users.Get(username); ok {
		msg.Origin.Send(events.RegisterFailedMessage("Another user with this username already exists."))
		return
	}

	details, err := parseUserDetails(msg.Args[1:])
	if err != nil {
		msg.Origin.Send(events.RegisterFailedMessage(err.Error()))
		return
	}
	password := []byte(msg.Args[1])
	server.startAuth(msg, ctx, func() (*models.User, error) {
		hash, err := server.Hash.GenerateHash(password)
		if err != nil {
			return nil, errors.New("could not process password")
		}
		return &models.User{
			Username: username,
			Password: *hash,
			Details:  details,
		}, nil
	})
}

func quit(_ *Server, msg events.ClientMessage, _ *commandContext) { _ = msg.Origin.Close() }

func ping(_ *Server, msg events.ClientMessage, _ *commandContext) {
	msg.Origin.Send(events.PongMessage())
}

func say(server *Server, msg events.ClientMessage, ctx *commandContext) {
	if len(msg.Args) != 1 || strings.TrimSpace(msg.Args[0]) == "" {
		// do not send an error since it's easy to accidentally press enter with an empty message
		return
	}

	channel, ok := server.Channels.Get(ctx.session.channelName)
	if !ok {
		log.Println("Could not find channel", ctx.session.channelName)
		msg.Origin.Send(events.ServerErrMessage("Failed to send message."))
		return
	}
	broadcastToChannel(
		server,
		channel,
		ctx.user.Username,
		FormatChannelMessage(ctx.session.channelName, ctx.user.Username, msg.Args[0]),
	)
}

func listFromNames(items map[string]bool) []string {
	names := make([]string, 0, len(items))
	for name := range items {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func chat(server *Server, msg events.ClientMessage, ctx *commandContext) {
	channels, err := server.Channels.List()
	if err != nil {
		log.Println("Error listing channels:", err)
		msg.Origin.Send(events.ServerErrMessage("Error listing channels."))
		channels = []string{}
	}
	slices.Sort(channels)
	msg.Origin.Send(events.ReportChannelsMessage(channels))
	msg.Origin.Send(events.BuddyListMessage(listFromNames(ctx.user.Buddies)))
	msg.Origin.Send(events.IgnoreListMessage(listFromNames(ctx.user.Ignored)))

	if ctx.session.channelName != "" {
		channel, ok := server.Channels.Get(ctx.session.channelName)
		if !ok {
			log.Println("Error getting channel:", ctx.session.channelName)
			msg.Origin.Send(events.ServerErrMessage("Error getting user list."))
			return
		}
		msg.Origin.Send(events.ReportUsersMessage(ctx.session.channelName, listFromNames(channel.Members)))
	}
}

func join(server *Server, msg events.ClientMessage, ctx *commandContext) {
	if len(msg.Args) < 1 {
		// TODO: do not do if the originator doesn't live on the server?
		msg.Origin.Send(events.ServerErrMessage("Missing channel name."))
		return
	}
	channel, ok := server.Channels.Get(msg.Args[0])
	if !ok {
		msg.Origin.Send(events.ServerErrMessage("Channel not found."))
		return
	}
	if ctx.session.channelName == channel.Name {
		msg.Origin.Send(events.ServerErrMessage("You already are in this channel."))
		return
	}

	if ctx.session.channelName != "" {
		oldChannel, ok := server.Channels.Get(msg.Args[0])
		if !ok {
			return
		}
		delete(oldChannel.Members, ctx.user.Username)
		err := server.Channels.Update(oldChannel)
		if err != nil {
			log.Println("Failed to update channel.", err)
			msg.Origin.Send(events.ServerErrMessage("Error while leaving channel."))
			oldChannel.Members[ctx.user.Username] = true
			return
		}
		AnnounceLeftChannel(server, oldChannel, ctx.user.Username)
		ctx.session.channelName = ""
	}
	channel.Members[ctx.user.Username] = true
	err := server.Channels.Update(channel)
	if err != nil {
		log.Println("Failed to update channel.", err)
		msg.Origin.Send(events.ServerErrMessage("Error while joining channel."))
		delete(channel.Members, ctx.user.Username)
		msg.Origin.Send(events.ServerMessage("Left channel"))
		return
	}

	ctx.session.channelName = channel.Name
	AnnounceJoinedChannel(server, channel, ctx.user.Username)
	msg.Origin.Send(events.ReportUsersMessage(channel.Name, listFromNames(channel.Members)))
}

func create(server *Server, msg events.ClientMessage, _ *commandContext) {
	if len(msg.Args) < 1 {
		msg.Origin.Send(events.ServerErrMessage("Missing channel name."))
		return
	}

	channels, err := server.Channels.List()
	if err != nil {
		log.Println("Failed to create channel", err)
		msg.Origin.Send(events.ServerErrMessage("Failed to create channel."))
		return
	}
	if len(channels) >= 256 {
		msg.Origin.Send(events.ServerErrMessage("Channel limit reached."))
		return
	}

	name := msg.Args[0]
	if !strings.HasPrefix(name, "#") {
		name = "#" + name
	}

	_, ok := server.Channels.Get(name)
	if ok {
		msg.Origin.Send(events.ServerErrMessage("Channel already exists."))
		return
	}

	if !validateName(strings.TrimPrefix(name, "#"), 63) {
		msg.Origin.Send(events.ServerErrMessage("Channel name must start with '#'."))
		return
	}

	err = server.Channels.Create(models.NewChannel(name))
	if err != nil {
		log.Println("Failed to create channel", err)
		msg.Origin.Send(events.ServerErrMessage("Failed to create channel."))
		return
	}
	msg.Origin.Send(events.ServerMessage(fmt.Sprintf("Created channel: %s", name)))
	for _, sess := range server.online {
		sess.client.Send(events.ChannelAddedMessage(name))
	}
}

func directMessage(server *Server, msg events.ClientMessage, ctx *commandContext) {
	if len(msg.Args) < 2 {
		msg.Origin.Send(events.ServerErrMessage("Expected a username."))
		return
	}
	if strings.TrimSpace(msg.Args[1]) == "" {
		// do not send an error since it's easy to accidentally press enter with an empty message
		return
	}

	target, ok := server.Users.Get(msg.Args[0])
	if !ok {
		msg.Origin.Send(events.ServerErrMessage("User not found."))
		return
	}

	if ctx.user.Ignored[target.Username] {
		msg.Origin.Send(events.ServerErrMessage("User is ignored."))
		return
	}

	msg.Origin.Send(FormatWhisperSender(target.Username, msg.Args[1]))

	targetSession, online := server.online[target.Username]
	if !online || targetSession.client.IsClosed() {
		msg.Origin.Send(events.ServerErrMessage("User is offline."))
		return
	}
	if target.Ignored[target.Username] {
		return
	}
	targetSession.client.Send(FormatWhisperTarget(ctx.user.Username, msg.Args[1]))
}

func whois(server *Server, msg events.ClientMessage, _ *commandContext) {
	if len(msg.Args) < 1 {
		msg.Origin.Send(events.ServerErrMessage("Expected a username."))
		return
	}
	user, ok := server.Users.Get(msg.Args[0])
	if !ok {
		msg.Origin.Send(events.ServerErrMessage("User not found."))
		return
	}

	msg.Origin.Send(FormatWhoisAnswer(user))
}

func newDetails(server *Server, msg events.ClientMessage, ctx *commandContext) {
	details, err := parseUserDetails(msg.Args)
	if err != nil {
		authError(msg.Origin, msg.Command, err.Error())
		return
	}
	username := ctx.user.Username
	password := []byte(msg.Args[0])
	server.startAuth(msg, ctx, func() (*models.User, error) {
		hash, err := server.Hash.GenerateHash(password)
		if err != nil {
			return nil, errors.New("could not process password")
		}
		return &models.User{
			Username: username,
			Password: *hash,
			Details:  details,
		}, nil
	})
}

func buddy(server *Server, msg events.ClientMessage, ctx *commandContext) {
	if len(msg.Args) < 2 {
		msg.Origin.Send(events.ServerErrMessage("Expected an operation and a username."))
		return
	}

	target := msg.Args[1]
	if _, ok := server.Users.Get(target); !ok {
		msg.Origin.Send(events.ServerErrMessage("User not found."))
		return
	}
	if target == ctx.user.Username {
		msg.Origin.Send(events.ServerErrMessage("Cannot add yourself as a buddy."))
		return
	}
	switch strings.ToUpper(msg.Args[0]) {
	case "ADD":
		if ctx.user.Buddies[target] {
			msg.Origin.Send(events.ServerErrMessage("User is already your buddy."))
			return
		}
		if len(ctx.user.Buddies) >= 128 {
			msg.Origin.Send(events.ServerErrMessage("Buddy limit reached."))
			return
		}
		ctx.user.Buddies[target] = true
		if !saveUser(server, msg, ctx.user) {
			return
		}
		msg.Origin.Send(events.BuddyAddMessage(target))
		msg.Origin.Send(events.BuddyStatusMessage(target, buddyStatus(server, target)))
		msg.Origin.Send(events.ServerMessage(fmt.Sprintf("%s is now your buddy!", target)))
	case "REMOVE":
		if !ctx.user.Buddies[target] {
			msg.Origin.Send(events.ServerErrMessage("User is not your buddy."))
			return
		}
		delete(ctx.user.Buddies, target)
		if !saveUser(server, msg, ctx.user) {
			return
		}
		msg.Origin.Send(events.BuddyRemoveMessage(target))
		msg.Origin.Send(events.ServerMessage(fmt.Sprintf("%s is not your buddy anymore :(", target)))
	default:
		msg.Origin.Send(events.ServerErrMessage("Unsupported operation."))
	}
}

func ignore(server *Server, msg events.ClientMessage, ctx *commandContext) {
	if len(msg.Args) < 2 {
		msg.Origin.Send(events.ServerErrMessage("Expected an operation and a username."))
		return
	}

	target := msg.Args[1]
	if _, exists := server.Users.Get(target); !exists {
		msg.Origin.Send(events.ServerErrMessage("User not found."))
		return
	}
	if target == ctx.user.Username {
		msg.Origin.Send(events.ServerErrMessage("Cannot ignore yourself."))
		return
	}
	switch strings.ToUpper(msg.Args[0]) {
	case "ADD":
		if ctx.user.Ignored[target] {
			msg.Origin.Send(events.ServerErrMessage("User is already ignored."))
			return
		}
		if len(ctx.user.Ignored) >= 128 {
			msg.Origin.Send(events.ServerErrMessage("Ignore limit reached."))
			return
		}
		ctx.user.Ignored[target] = true
		if !saveUser(server, msg, ctx.user) {
			return
		}
		msg.Origin.Send(events.IgnoreAddMessage(target))
		msg.Origin.Send(events.ServerMessage(fmt.Sprintf("%s is now ignored.", target)))
	case "REMOVE":
		if !ctx.user.Ignored[target] {
			msg.Origin.Send(events.ServerErrMessage("User is not ignored."))
			return
		}
		delete(ctx.user.Ignored, target)
		if !saveUser(server, msg, ctx.user) {
			return
		}
		msg.Origin.Send(events.IgnoreRemoveMessage(target))
		msg.Origin.Send(events.ServerMessage(fmt.Sprintf("%s is not ignored anymore.", target)))
	default:
		msg.Origin.Send(events.ServerErrMessage("Unsupported operation."))
	}
}

func report(_ *Server, msg events.ClientMessage, ctx *commandContext) {
	if len(msg.Args) != 1 {
		msg.Origin.Send(events.ServerErrMessage("Expected one username."))
		return
	}

	log.Printf("REPORT: user %q reported %q", ctx.user.Username, msg.Args[0])
	msg.Origin.Send(events.ServerMessage(fmt.Sprintf("%s successfully reported.", msg.Args[0])))
}
