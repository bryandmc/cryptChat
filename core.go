package cryptchat

import (
	"net"
)

// Structs / data types

// Message is a type meant to encapsulate a message from a user to a user or room
type Message struct {
	msg        string // might switch to []byte if it proves easier with encryption
	attachment []byte
	sentFrom   *User
	sentTo     *User
	room       *Room
	isToRoom   bool
}

func (m *Message) Send() {
	if m.isToRoom {
		log.Debug(len(m.room.users))
		for _, u := range m.room.users {
			log.Info("Sending to:", u.name)
			u.channel <- m
		}
	} else {
		m.sentTo.channel <- m
	}
}

type CommandType uint8

// This emulates an 'enum' type structure
const (
	SEND_DIRECT CommandType = iota
	SEND_ROOM
	JOIN_ROOM
	CREATE_ROOM
	REMOVE_ROOM
)

type Argument map[string]string

// Command is the basic structure used to parse and then respond to user
// commands. They are parsed from raw input.
type Command struct {
	Cmd  CommandType
	Args []Argument // key:val mapped list of arguments
	Msg  *Message
}

type User struct {
	name    string
	conn    *net.Conn
	channel chan *Message
	online  bool
}

type Room struct {
	name     string
	users    []*User
	messages []*Message
}
