package events

import (
	"fmt"
	"strings"
)

type Message struct {
	Origin  *Client
	Raw     string
	Command string
	Args    []string
}

func NewMessageFromRaw(raw string) Message {
	args := strings.Fields(raw)
	if len(args) == 0 {
		return Message{Raw: raw, Command: raw}
	}
	return Message{
		Raw:     raw,
		Command: strings.ToUpper(args[0]),
		Args:    args[1:],
	}
}

func NewMessageFromClient(raw string, client *Client) Message {
	msg := NewMessageFromRaw(raw)
	msg.Origin = client
	return msg
}

func (msg *Message) String() string {
	return fmt.Sprint(msg.Raw)
}

var (
	QuitCommand           = "QUIT"
	PingCommand           = "PING"
	PongCommand           = "PONG"
	RegisterCommand       = "REGISTER"
	RegisterOkCommand     = "REGISTER_OK"
	RegisterFailedCommand = "REGISTER_FAIL"
	LoginCommand          = "LOGIN"
	LoginOkCommand        = "LOGIN_OK"
	LoginFailedCommand    = "LOGIN_FAIL"
	DisplayMessageCommand = "DISPLAY_MSG"
	ServerErrorCommand    = "SERVER_ERR"
	ClientProtocolCommand = "CLIENT_PROTOCOL"
	ClientTypeCommand     = "CLIENT_TYPE"
	SayCommand            = "SAY"
	UserMessageCommand    = "USER_MSG"
)

var (
	PingMessage       = NewMessageFromRaw(PingCommand)
	PongMessage       = NewMessageFromRaw(PongCommand)
	RegisterOkMessage = NewMessageFromRaw(RegisterOkCommand)
	LoginOkMessage    = NewMessageFromRaw(LoginOkCommand)
)

func RegisterFailedMessage(reason string) Message {
	return NewMessageFromRaw(RegisterFailedCommand + " " + reason)
}

func LoginFailedMessage(reason string) Message {
	return NewMessageFromRaw(LoginFailedCommand + " " + reason)
}

func DisplayMessage(msg string) Message {
	return NewMessageFromRaw(DisplayMessageCommand + " " + msg)
}

func ServerErrMessage(err string) Message {
	return NewMessageFromRaw(ServerErrorCommand + " " + err)
}

func UserMessage(username string, msg string) Message {
	return NewMessageFromRaw(fmt.Sprintf("%s %s %s", UserMessageCommand, username, msg))
}
