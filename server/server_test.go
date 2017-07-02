package main

import (
	"net"
	"reflect"
	"testing"
)

func TestCreateUser(t *testing.T) {
	connection := new(net.Conn)
	name := "Bryan"
	type args struct {
		username string
		c        *net.Conn
	}
	tests := []struct {
		name string
		args args
		want User
	}{
		{name: "basic create user",
			args: args{
				username: name,
				c:        connection,
			},
			want: User{
				Name:    name,
				conn:    connection,
				channel: make(chan *Message),
			},
		},
	}
	// Had to customize this one
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateUser(tt.args.username, tt.args.c); !(got.conn == tt.args.c && got.Name == tt.args.username) && users[tt.args.username] == got {
				t.Errorf("CreateUser() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCreateRoom(t *testing.T) {
	type args struct {
		roomname string
	}
	tests := []struct {
		name string
		args args
		want Room
	}{
		{
			name: "basic create room",
			args: args{
				roomname: "test room",
			},
			want: Room{
				Name:     "test room",
				Users:    []*User{},
				Messages: []*Message{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateRoom(tt.args.roomname); !reflect.DeepEqual(*got, tt.want) && rooms[tt.args.roomname] != got {
				t.Errorf("CreateRoom() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRemoveRoom(t *testing.T) {
	CreateRoom("room test")
	type args struct {
		roomname string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "basic",
			args: args{
				roomname: "room test",
			},
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			RemoveRoom(tt.args.roomname)
		})
	}
}

func TestJoinRoom(t *testing.T) {
	type args struct {
		usr *User
		rm  *Room
	}
	tests := []struct {
		name string
		args args
	}{
	// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			JoinRoom(tt.args.usr, tt.args.rm)
		})
	}
}
