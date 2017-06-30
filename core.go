package cryptchat

import (
	"fmt"
	"net"
)

// Structs / data types

// Message is a type meant to encapsulate a message from a user to a user or room
type Message struct {
	msg        string
	attachment []byte
	sentFrom   *User
	sentTo     *User
	room       *Room
	isToRoom   bool
}

func (m *Message) Send() {
	if m.isToRoom {
		for count, u := range m.room.users {
			fmt.Println("User count:", count)
			u.channel <- m
		}
	} else {
		m.sentTo.channel <- m
	}
}

type User struct {
	name    string
	conn    *net.Conn
	channel chan *Message
}

func (u *User) IsOnline() bool {
	return false
}

type Room struct {
	name     string
	users    []*User
	messages []*Message
}
