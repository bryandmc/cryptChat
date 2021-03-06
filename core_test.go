package cryptchat

import (
	"net"
	"testing"
	"time"
)

func TestMessage_Send(t *testing.T) {
	go Listen(ReadHandler)
	timeWait, _ := time.ParseDuration("10s")
	c, _ := net.DialTimeout("tcp", "localhost:1234", timeWait)
	usr := CreateUser("Bryan", &c)
	rm := CreateRoom("test send room")
	JoinRoom(usr, rm)
	quit := make(chan bool)

	// TODO -- there is an error if you attempt
	// to test after running the program beforehand.
	// Running test again right after works, however.
	// Unsure as to why this is the case. Must have something to
	// do with 'go Listen(ReadHandler)' and the program not
	// closing properly from a running state.
	go RecieveMsgs(usr, &quit)

	type fields struct {
		msg        []byte
		attachment []byte
		sentFrom   *User
		sentTo     *User
		room       *Room
		isToRoom   bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "basic user",
			fields: fields{
				msg:        []byte("Hey!!"),
				attachment: *new([]byte),
				sentFrom:   new(User),
				sentTo:     usr,
				room:       nil,
				isToRoom:   false,
			},
		},
		{
			name: "basic room",
			fields: fields{
				msg:        []byte("Hey!!"),
				attachment: *new([]byte),
				sentFrom:   usr,
				sentTo:     nil,
				room:       rm,
				isToRoom:   true,
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				Body:       tt.fields.msg,
				Attachment: tt.fields.attachment,
				SentFrom:   tt.fields.sentFrom,
				SentTo:     tt.fields.sentTo,
				Room:       tt.fields.room,
				IsToRoom:   tt.fields.isToRoom,
			}
			m.Send()
		})
	}
	quit <- true
}
