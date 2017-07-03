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
	log.Notice("CONNECT")
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

func ReadInput(c *net.Conn, out chan string) {

	for {
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		out <- text
		log.Notice("read input")
	}
}

func WriteOutput(c *net.Conn, in chan []byte) {
	for {
		w := <-in
		log.Notice("write output NET")
		(*c).Write(w)
	}
}

// 	for {
// 		if err != nil {
// 			fmt.Println(err.Error())
// 		}
// 		b := make([]byte, 1024*10)
// 		conn.Read(b)
// 		col := color.New(color.FgBlack).Add(color.BgCyan)
// 		col.Print(string(b))
// 		reader := bufio.NewReader(os.Stdin)
// 		text, _ := reader.ReadString('\n')
// 		conn.Write([]byte(text))
// 	}
// }

// ParseInput takes the input from a channel and parses it into a Command structure
// which then produces a 'Comnmand' type that is sent to the output channel.
func ParseInput(in chan string, out chan Command) {

	for {
		s := <-in
		log.Notice("parse input")
		re := regexp.MustCompile(`(.*)(\<|\>|\|)(.*)`) // end at newline
		stripNewline := regexp.MustCompile("\n")
		index := stripNewline.FindStringIndex(s) // gives it in [start, end] []int format

		cmd := &Command{
			Cmd:  SEND_DIRECT, //default option
			Args: []Argument{},
			Msg:  &Message{},
		}
		splitInput := re.FindStringSubmatch(s[:index[0]])
		log.Notice(re)
		for c, v := range splitInput {
			switch c {
			case 0: //full match
				log.Warning(c, strings.TrimSpace(v), " FULL")
			case 1: //left
				log.Warning(c, strings.TrimSpace(v), " LEFT")
			case 2: //operator
				userLock.Lock()
				log.Warning(c, strings.TrimSpace(v), " OPERATOR")
				if v == ">" {

					*(cmd.Msg) = Message{
						Body:   splitInput[c-1],        // the left is the message
						SentTo: users[splitInput[c+1]], // the right is the user.
					}
				} else if v == "<" {
					*(cmd.Msg) = Message{
						Body:   splitInput[c+1],        // the right is the message
						SentTo: users[splitInput[c-1]], // the left is the user.
					}
				}
				userLock.Unlock()
			case 3: //right
				log.Warning(c, strings.TrimSpace(v), " RIGHT")
			}
		}
		log.Error(cmd)
		log.Info("Index position of newline character:", index)
		out <- *cmd
	}

}

func EncryptMessage(in chan Command, out chan Command) {
	for {
		cmd := <-in
		log.Notice("encrypt input")
		b, err := Encrypt([]byte(cmd.Msg.Body), []byte("Whatever I want as a key and it'll work just fine"[:32]))
		if err != nil {
			log.Critical(err.Error())
		} else { // only pass it on if there is no error
			log.Debug(b)
			cmd.Msg.Body = string(b)
			out <- cmd
		}
	}
}

func MarshalMessage(in chan Command, out chan []byte) {
	for {
		cmd := <-in
		log.Notice("Marshal input")
		marshMessage, err := json.Marshal(cmd)
		if err != nil {
			log.Critical(err.Error())
		} else {
			out <- marshMessage
		}
	}
}
