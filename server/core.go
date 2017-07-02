package main

import (
	"net"
)

// Structs / data types

// Message is a type meant to encapsulate a message from a user to a user or room
type Message struct {
	Msg        string `json:"msg,omitempty"` // might switch to []byte if it proves easier with encryption
	Attachment []byte `json:"attachment,omitempty"`
	SentFrom   *User  `json:"sent_from,omitempty"`
	SentTo     *User  `json:"sent_to,omitempty"`
	Room       *Room  `json:"room,omitempty"`
	IsToRoom   bool   `json:"is_to_room,omitempty"`
}

// Send is used to send a message within the system. It follows a simple rule
// for sending messages. Either single user --> single user or
// single user --> room (many users).
func (m *Message) Send() {
	if m.IsToRoom {
		log.Debug(len(m.Room.Users))
		for _, u := range m.Room.Users {
			log.Info("Sending to:", u.Name)
			u.channel <- m
		}
	} else {
		m.SentTo.channel <- m
	}
}

// CommandType is an alias for unsinged 8-bit int for use in creating
// the command enum below.
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
	Cmd  CommandType `json:"cmd,omitempty"`
	Args []Argument  `json:"args,omitempty"` // key:val mapped list of arguments
	Msg  *Message    `json:"msg,omitempty"`
}

type User struct {
	Name    string `json:"name,omitempty"`
	conn    *net.Conn
	channel chan *Message
	Online  bool `json:"online,omitempty"`
}

type Room struct {
	Name     string     `json:"name,omitempty"`
	Users    []*User    `json:"users,omitempty"`
	Messages []*Message `json:"messages,omitempty"`
}