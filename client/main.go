package main

import cc "github.com/bryandmc/cryptchat"

func main() {
	c, err := cc.Connect("localhost:1234")
	if err != nil {
		return
	}
	done := make(chan bool)
	// read input from user
	readString := make(chan string)
	go cc.ReadInput(c, readString)

	// pass user input to parsing function
	cmdOut := make(chan cc.Command)
	go cc.ParseInput(readString, cmdOut)

	// pass to message encryption function
	encryptOut := make(chan cc.Command)
	go cc.EncryptMessage(cmdOut, encryptOut)

	// pass to json Marshaler
	marshalOut := make(chan []byte)
	go cc.MarshalMessage(encryptOut, marshalOut)

	// pass to network Writer
	go cc.WriteOutput(c, marshalOut)

	// wait group so that all goroutines finish before exiting
	<-done // just used to block currently
}
