package events

import (
	"strings"
)

type Message struct {
	Origin  *Client
	Raw     string
	Command Command
	Args    []string
}

func NewMessageFromRaw(raw string) Message {
	args := strings.Fields(raw)
	if len(args) == 0 {
		return Message{}
	}
	return Message{
		Raw:     raw,
		Command: Command(strings.ToUpper(args[0])),
		Args:    args[1:],
	}
}

func NewMessageFromClient(raw string, client *Client) Message {
	msg := NewMessageFromRaw(raw)
	msg.Origin = client
	return msg
}

func NewOutboundMessage(cmd Command, args ...string) Message {
	raw := string(cmd)
	if len(args) > 0 {
		raw += " " + strings.Join(args, " ")
	}
	return Message{
		Raw:     raw,
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
	return NewOutboundMessage(ReportChannelsCommand, channels...)
}

func ReportUsersMessage(users []string) Message {
	return NewOutboundMessage(ReportUsersCommand, users...)
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
