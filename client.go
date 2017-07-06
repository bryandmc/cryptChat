package cryptchat

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"encoding/json"

	"github.com/fatih/color"
)

var user User
var key = []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"[:32])

func printBanner() {
	fmt.Println()
	color.Blue("##############################################")
	color.Yellow("\tWelcome to cryptchat.")
	color.Blue("##############################################")
	fmt.Println()
}

// Connect establishes the initial connection to the server and returns
// the connection, error. This is a wrapper to handle errors.
func Connect(host string) (*net.Conn, error) {
	printBanner()
	t, err := time.ParseDuration("10s")
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	conn, err := net.DialTimeout("tcp", host, t)
	if err != nil {
		log.Debug(err.Error())
		return nil, err
	}
	return &conn, nil
}

func SendUserName() *Command {
	// send personal username to server
	fmt.Print("Enter your username: ")
	reader := bufio.NewReader(os.Stdin)
	username, _ := reader.ReadString('\n')
	userCmd := &Command{
		Cmd: SEND_USERNAME,
		Args: Arguments{
			"connect_username": strings.TrimSpace(username),
		},
		Msg: &Message{},
	}
	var currentUser = User{
		Name:   username,
		Online: true,
	}
	user = currentUser
	return userCmd
}

func ReadInput(c *net.Conn, out chan string) {
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		out <- text
	}
}

func WriteOutput(c *net.Conn, in chan []byte) {
	for {
		w := <-in
		(*c).Write(w)
	}
}

// ParseInput takes the input from a channel and parses it into a Command structure
// which then produces a 'Comnmand' type that is sent to the output channel.
func ParseInput(in chan string, out chan Command) {
	for {
		s := <-in
		re := regexp.MustCompile(`(.*)(\<|\>|\|)(.*)`) // end at newline
		stripNewline := regexp.MustCompile("\n")
		index := stripNewline.FindStringIndex(s) // gives it in [start, end] []int format

		cmd := &Command{
			Cmd:  SEND_DIRECT, //default option
			Args: Arguments{},
			Msg:  &Message{},
		}
		splitInput := re.FindStringSubmatch(s[:index[0]])
		for c, v := range splitInput {
			switch c { // (0, 1, ..., 3) currently unused. Left here because they seem like they might get used soon
			case 0: //full match
			case 1: //left
			case 2: //operator
				if v == ">" {
					*(cmd.Msg) = Message{
						Body: []byte(strings.TrimSpace(splitInput[c-1])), // the left is the message
					}
					cmd.Args = Arguments{
						"to_username":   strings.TrimSpace(splitInput[c+1]),
						"from_username": user.Name,
					}
				} else if v == "<" {
					*(cmd.Msg) = Message{
						Body: []byte(strings.TrimSpace(splitInput[c+1])), // the right is the message
					}
					cmd.Args = Arguments{
						"to_username":   strings.TrimSpace(splitInput[c-1]),
						"from_username": user.Name,
					}
				} else if v == "|" {
					*(cmd.Msg) = Message{
						Body: []byte(strings.TrimSpace(splitInput[c-1])), // the left is the message
					}
					cmd.Args = Arguments{
						"to_room":       strings.TrimSpace(splitInput[c+1]),
						"from_username": user.Name,
					}
				}
			case 3: //right
			}
		}
		out <- *cmd
	}

}

func EncryptMessage(in chan Command, out chan Command) {
	for {
		cmd := <-in
		b, err := Encrypt([]byte(cmd.Msg.Body), key)
		if err != nil {
			log.Critical(err.Error())
		} else { // only pass it on if there is no error
			cmd.Msg.Body = b // bypassing encryption temporarily
			out <- cmd
		}
	}
}

func ListenResponse(c *net.Conn) {
	buff := make([]byte, 1024*4)
	for {
		count, err := (*c).Read(buff)
		if err != nil {
			fmt.Println(err)
		}
		if count <= 0 {
			return
		}
		decryptedText, err := Decrypt(buff[:count], key)
		if err != nil {
			fmt.Println(err.Error())
		} else {
			fmt.Println(string(decryptedText))
		}
	}
}

func MarshalMessage(in chan Command, out chan []byte) {
	for {
		cmd := <-in
		marshMessage, err := json.Marshal(cmd)
		if err != nil {
			log.Critical(err.Error())
		} else {
			out <- marshMessage
		}
	}
}
