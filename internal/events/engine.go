package events

// NOTE: This design won't last the test of time
//  and doesn't scale neatly since every message is
//  broadcasted to every server in the cluster
//  a real pub/sub design should be implemented
//  however to the current scale of this project
//  its faster and easier to use this design.
//  Eventually the pub/sub design with proper
//  routing can be implemented at driver level

type Engine interface {
	Send(msg Message) error
	Receive() <-chan Message
}
