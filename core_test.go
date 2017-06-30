package cryptchat

import (
	"net"
	"testing"
	"time"
)

func TestMessage_Send(t *testing.T) {
	//go Listen(ReadHandler)
	timeWait, _ := time.ParseDuration("10s")
	c, _ := net.DialTimeout("tcp", "localhost:1234", timeWait)
	usr := CreateUser("Bryan", &c)
	rm := CreateRoom("test send room")
	JoinRoom(&usr, &rm)
	go RecieveMsgs(&usr)
	type fields struct {
		msg        string
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
				msg:        "Hey!!",
				attachment: *new([]byte),
				sentFrom:   new(User),
				sentTo:     &usr,
				room:       nil,
				isToRoom:   false,
			},
		},
		{
			name: "basic room",
			fields: fields{
				msg:        "Hey!!",
				attachment: *new([]byte),
				sentFrom:   &usr,
				sentTo:     nil,
				room:       &rm,
				isToRoom:   true,
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				msg:        tt.fields.msg,
				attachment: tt.fields.attachment,
				sentFrom:   tt.fields.sentFrom,
				sentTo:     tt.fields.sentTo,
				room:       tt.fields.room,
				isToRoom:   tt.fields.isToRoom,
			}
			m.Send()
		})
	}
}
