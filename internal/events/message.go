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
	UserJoinCommand       Command = "USER_JOIN"
	UserLeaveCommand      Command = "USER_LEAVE"
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

func UserMessage(username string, msg string) Message {
	return NewOutboundMessage(UserMessageCommand, username, msg)
}
