package events

import (
	"strings"
)

const whitespace = " \t"

var sanitizer = strings.NewReplacer("\r", " ", "\n", " ")

type Message struct {
	Raw     string
	Command Command
	Args    []string
}

type ClientMessage struct {
	Message
	Origin *Client
}

func popArg(text string) (string, string) {
	if i := strings.IndexAny(text, whitespace); i >= 0 {
		return text[:i], strings.TrimLeft(text[i+1:], whitespace)
	}
	return text, ""
}

func NewMessageFromRaw(raw string) Message {
	rawTrimmed := strings.TrimLeft(raw, " \t")
	if rawTrimmed == "" {
		return Message{}
	}
	command, rest := popArg(rawTrimmed)
	cmd := Command(strings.ToUpper(command))
	var args []string
	switch cmd {
	case SayCommand:
		args = []string{rest}
	case DirectMessageCommand:
		target, body := popArg(rest)
		args = []string{target, body}
	default:
		args = strings.Fields(rest)
	}
	return Message{
		Raw:     raw,
		Command: cmd,
		Args:    args,
	}
}

func NewMessageFromClient(raw string, client *Client) ClientMessage {
	return ClientMessage{
		Message: NewMessageFromRaw(raw),
		Origin:  client,
	}
}

func NewOutboundMessage(cmd Command, args ...string) Message {
	raw := string(cmd) + " " + sanitizer.Replace(strings.Join(args, " "))
	return Message{
		Raw:     strings.TrimSpace(raw),
		Command: cmd,
		Args:    args,
	}
}

func (msg *Message) String() string {
	return msg.Raw
}

type Command string

const (
	QuitCommand           Command = "QUIT"
	PingCommand           Command = "PING"
	PongCommand           Command = "PONG"
	RegisterCommand       Command = "REGISTER"
	RegisterOkCommand     Command = "REGISTER_OK"
	RegisterFailedCommand Command = "REGISTER_FAIL"
	LoginCommand          Command = "LOGIN"
	LoginOkCommand        Command = "LOGIN_OK"
	LoginFailedCommand    Command = "LOGIN_FAIL"
	DisplayMessageCommand Command = "DISPLAY_MSG"
	ServerErrorCommand    Command = "SERVER_ERR"
	ClientProtocolCommand Command = "CLIENT_PROTOCOL"
	ClientTypeCommand     Command = "CLIENT_TYPE"
	SayCommand            Command = "SAY"
	UserMessageCommand    Command = "USER_MSG"
	ChatCommand           Command = "CHAT"
	JoinCommand           Command = "JOIN"
	DirectMessageCommand  Command = "MSG"
	CreateChannelCommand  Command = "CREATE"
	WhoisCommand          Command = "WHOIS"
	NewDetailsCommand     Command = "NEWDETAILS"
	ChannelAddedCommand   Command = "CHANNEL_ADDED"
	ChannelRemovedCommand Command = "CHANNEL_REMOVED"
	ReportChannelsCommand Command = "REPORT_CHANNELS"
	UserJoinCommand       Command = "USER_JOIN"
	UserLeaveCommand      Command = "USER_LEAVE"
	ReportUsersCommand    Command = "REPORT_USERS"
	ServerMessageCommand  Command = "SERVER_MSG"
	NewDetailsOkCommand   Command = "NEWDETAILS_OK"
	UserDetailsCommand    Command = "USER_DETAILS"
	BuddyStatusCommand    Command = "BUDDY_STATUS"
	BuddyListCommand      Command = "BUDDY_LIST"
	BuddyAddCommand       Command = "BUDDY_ADD"
	BuddyRemoveCommand    Command = "BUDDY_DEL"
	BuddyCommand          Command = "BUDDY"
	IgnoreListCommand     Command = "IGNORE_LIST"
	IgnoreAddCommand      Command = "IGNORE_ADD"
	IgnoreRemoveCommand   Command = "IGNORE_DEL"
	IgnoreCommand         Command = "IGNORE"
	ReportCommand         Command = "REPORT"
)

func PingMessage() Message {
	return NewOutboundMessage(PingCommand)
}

func PongMessage() Message {
	return NewOutboundMessage(PongCommand)
}

func RegisterOkMessage() Message {
	return NewOutboundMessage(RegisterOkCommand)
}

func LoginOkMessage() Message {
	return NewOutboundMessage(LoginOkCommand)
}

func NewDetailsOkMessage() Message {
	return NewOutboundMessage(NewDetailsOkCommand)
}

func RegisterFailedMessage(reason string) Message {
	return NewOutboundMessage(RegisterFailedCommand, reason)
}

func LoginFailedMessage(reason string) Message {
	return NewOutboundMessage(LoginFailedCommand, reason)
}

func DisplayMessage(msg string) Message {
	return NewOutboundMessage(DisplayMessageCommand, msg)
}

func ServerErrMessage(err string) Message {
	return NewOutboundMessage(ServerErrorCommand, err)
}

func ServerMessage(msg string) Message {
	return NewOutboundMessage(ServerMessageCommand, msg)
}

func UserMessage(msg string) Message {
	// It's the server responsibility to format the message
	//  The client will "dumb print" whatever is sent
	//  So the message should be like `[#channel] user: message`
	//  Or [@user] hello

	// Also, any message the user sent must be displayed to them
	//  in the same fashion

	// This is not intended for direct use,
	//  use and create formatters in helpers.go
	return NewOutboundMessage(UserMessageCommand, msg)
}

func ChannelAddedMessage(channel string) Message {
	return NewOutboundMessage(ChannelAddedCommand, channel)
}

func ReportChannelsMessage(channels []string) Message {
	return NewOutboundMessage(ReportChannelsCommand, strings.Join(channels, " "))
}

func ReportUsersMessage(channel string, users []string) Message {
	return NewOutboundMessage(ReportUsersCommand, channel, strings.Join(users, " "))
}

func BuddyListMessage(buddies []string) Message {
	return NewOutboundMessage(BuddyListCommand, strings.Join(buddies, " "))
}

func IgnoreListMessage(ignored []string) Message {
	return NewOutboundMessage(IgnoreListCommand, strings.Join(ignored, " "))
}

func UserJoinMessage(username string) Message {
	return NewOutboundMessage(UserJoinCommand, username)
}

func UserLeaveMessage(username string) Message {
	return NewOutboundMessage(UserLeaveCommand, username)
}

func UserDetailsMessage(username string, args []string) Message {
	return NewOutboundMessage(UserDetailsCommand, username, strings.Join(args, " "))
}

func BuddyStatusMessage(username string, status BuddyStatus) Message {
	return NewOutboundMessage(BuddyStatusCommand, username, string(status))
}

func BuddyAddMessage(username string) Message {
	return NewOutboundMessage(BuddyAddCommand, username)
}

func BuddyRemoveMessage(username string) Message {
	return NewOutboundMessage(BuddyRemoveCommand, username)
}

func IgnoreAddMessage(username string) Message {
	return NewOutboundMessage(IgnoreAddCommand, username)
}

func IgnoreRemoveMessage(username string) Message {
	return NewOutboundMessage(IgnoreRemoveCommand, username)
}

type BuddyStatus string

const (
	BuddyStatusOnline  BuddyStatus = "+"
	BuddyStatusOffline BuddyStatus = "-"
)
