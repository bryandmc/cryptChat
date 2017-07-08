package cryptchat

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"regexp"
	"strings"
	"time"

	"encoding/json"

	"github.com/fatih/color"
)

var user User
var key []byte

type Config struct {
	CryptoKey string `json:"crypto_key,omitempty"`
}

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

func ReadConfig() {
	dat, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println(err.Error())
	}
	conf := new(Config)
	err = json.Unmarshal(dat, conf)
	if err != nil {
		fmt.Println("Marshalling error:", err.Error())
	} else {
		key = []byte(conf.CryptoKey[:32])
	}
}

// SendUserName is the first command that establishes a username for the user with
// the server.
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

// ReadInput is meant too be a small goroutine that is spawned, waiting for the moment
// when user input is typed and enter is pressed. Then it goes and waits again.
func ReadInput(c *net.Conn, out chan string) {
	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		out <- text
	}
}

// WriteOutput is a simple function that is meant to be a goroutine that simple waits for data
// to write to the socket that comes in on a channel.
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
		if len(splitInput) > 2 {
			if splitInput[2] == ">" {
				*(cmd.Msg) = Message{
					Body: []byte(strings.TrimSpace(splitInput[1])), // the left is the message
				}
				cmd.Args = Arguments{
					"to_username":   strings.TrimSpace(splitInput[3]),
					"from_username": user.Name,
				}
			} else if splitInput[2] == "<" {
				*(cmd.Msg) = Message{
					Body: []byte(strings.TrimSpace(splitInput[3])), // the right is the message
				}
				cmd.Args = Arguments{
					"to_username":   strings.TrimSpace(splitInput[1]),
					"from_username": user.Name,
				}
			} else if splitInput[2] == "|" {
				*(cmd.Msg) = Message{
					Body:     []byte(strings.TrimSpace(splitInput[1])), // the left is the message
					IsToRoom: true,
				}
				cmd.Args = Arguments{
					"to_room":       strings.TrimSpace(splitInput[3]),
					"from_username": user.Name,
				}
			}
		}
		out <- *cmd
	}

}

// EncryptMessage is used (like most other functions) as a small goroutine that will wait for
// data to be provided to it, so that it can Encrypt the message and pass it along.
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

// ListenResponse is the function that waits for information to be
// sent over the socket and then displays it to the user.
func ListenResponse(c *net.Conn) {
	buff := make([]byte, 1024)
	//buff := new(bytes.Buffer)
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
			fmt.Println(TimeResponse(), string(decryptedText))
		}
	}
}

// MarshalMessage is used to convert a Command struct into a json representation
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

// UnMarshalMessage is for converting messages from json --> Command structs
func UnMarshalMessage(in chan []byte, out chan Command) {
	for {
		cmd := Command{}
		inputData := <-in
		err := json.Unmarshal(inputData, &cmd)
		if err != nil {
			log.Error("Unmarshal: ", err.Error())
		} else {
			out <- cmd
		}
	}
}
