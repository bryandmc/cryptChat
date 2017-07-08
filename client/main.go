package main

import (
	cc "github.com/bryandmc/cryptchat"
)

func main() {
	cc.ReadConfig()
	c, err := cc.Connect("localhost:1234")
	if err != nil {
		return
	}

	done := make(chan bool)
	readString := make(chan string, 20)
	cmdOut := make(chan cc.Command, 20)
	encryptOut := make(chan cc.Command, 20)
	marshalOut := make(chan []byte, 20)
	userCmd := cc.SendUserName()
	cmdOut <- *userCmd //send username
	// read input from user
	go cc.ListenResponse(c)
	go cc.ReadInput(c, readString)

	// pass user input to parsing function
	go cc.ParseInput(readString, cmdOut)

	// pass to message encryption function
	go cc.EncryptMessage(cmdOut, encryptOut)

	// pass to json Marshaler
	go cc.MarshalMessage(encryptOut, marshalOut) //modified for testing !! first shold be (encryptOut, ??)

	// pass to network Writer
	go cc.WriteOutput(c, marshalOut)

	// wait group so that all goroutines finish before exiting
	<-done // just used to block currently
}
