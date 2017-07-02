package cryptchat

import (
	"net"
	"os"

	"regexp"

	"time"

	"sync"

	"strings"

	logging "github.com/op/go-logging"
)

// Users contains a list of the currently connected User objects
var users = make(map[string]*User)
var userLock = &sync.RWMutex{}

// Rooms contains a list of the currently available/active Room objects
var rooms = make(map[string]*Room)
var roomLock = &sync.RWMutex{}

var log = logging.MustGetLogger("server")
var format = logging.MustStringFormatter(
	`%{color}%{time:15:04:05.000} %{shortfunc} ▶ %{level:.4s} %{id:03x}%{color:reset} %{message}`,
)

// Listen listens for incoming connections then hands it off to another goroutine for handling
// gotta love go ;)
func Listen(handle func(*net.Conn)) {
	conn, err := net.Listen("tcp", "localhost:1234")
	if err != nil {
		log.Critical(err.Error())
	} else {
		for { // keeps accepting connections forever
			c, err := conn.Accept()
			if err != nil {
				log.Critical(err.Error())
			} else { // only want to handle if there WASN'T an error
				go handle(&c) // handle that shit in a goroutine so it's not blocking
			}
		}
	}
}

// RecieveMsgs blocks until it gets a message, sends that the the user and
// repeats. Nice and simple. Send on channel
func RecieveMsgs(usr *User, quit *chan bool) {
	for {
		select {
		case val := <-usr.channel:
			c := *usr.conn
			c.Write([]byte("[" + val.sentFrom.name + "] " + val.msg + "\n"))
		case <-(*quit):
			log.Debug("Quitting goroutine!")
			return
		}
	}
}

// ReadHandler handles all the incoming connections and reading from the
// socket. Little more complicated than I'd like currently.
var ReadHandler = func(c *net.Conn) {
	quit := make(chan bool)
	usr := CreateUser("Bryan", c)
	log.Debug(usr.name)

	// recieve messages
	go RecieveMsgs(usr, &quit)

	// create and join room
	rm := CreateRoom("test")
	JoinRoom(usr, rm)
	log.Debug(rm)

	// main read loop
	log.Debug("Starting goroutine to handle connection from:", (*c).RemoteAddr())
	for {
		count, buff, _ := readInput(c)
		if count <= 0 {
			log.Warningf("No data read (%d) closing goroutine.", count)
			quit <- true
			RemoveUser(usr.name) // this would work once we dissalow duplicate usernames
			return               // get out of here
		}

		// parse input
		cmd, err := ParseInput(buff, count)
		if err != nil {
			log.Critical(err.Error())
		}
		log.Info(cmd)

		s := string(buff[:count]) // convert to string regex
		re := regexp.MustCompile("\n")
		index := re.FindStringIndex(s) // gives it in [start, end] []int format
		log.Info("Index position of newline character:", index)
		// send messages here
		var msg = Message{
			msg:        s[:index[0]],
			attachment: []byte{},
			sentFrom:   usr,
			sentTo:     nil,
			room:       rm,
			isToRoom:   true,
		}
		msg.Send()
		//log.Critical(s[:index[0]])
	}
}

// ParseInput takes the input and parses it into a Command structure
// which tells the caller what the user wants to do. Error is returns based
// on errors in this process.
func ParseInput(input []byte, count int) (*Command, error) {
	s := string(input[:count])                     // convert to string regex
	re := regexp.MustCompile(`(.*)(\<|\>|\|)(.*)`) // end at newline
	stripNewline := regexp.MustCompile("\n")
	index := stripNewline.FindStringIndex(s) // gives it in [start, end] []int format

	cmd := &Command{
		Cmd:  SEND_DIRECT,
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
			log.Warning(c, strings.TrimSpace(v), " OPERATOR")
			if v == ">" {
				cmd.Cmd = SEND_DIRECT
				*(cmd.Msg) = Message{
					msg:    splitInput[c-1],        // the left is the message
					sentTo: users[splitInput[c+1]], // the right is the user. TODO: lock mutex
				}
			}
		case 3: //right
			log.Warning(c, strings.TrimSpace(v), " RIGHT")
		}

	}
	log.Error(cmd)
	log.Info("Index position of newline character:", index)
	return cmd, nil
}

// JoinRoom adds a user to a room.
func JoinRoom(usr *User, rm *Room) {
	roomLock.Lock()
	for _, val := range rm.users { // check for existing user
		if usr == val {
			log.Info(val)
			return
		}
	}
	rm.users = append(rm.users, usr)
	roomLock.Unlock()
}

// CreateRoom handles creating a new Room to be inserted into Rooms
func CreateRoom(roomname string) *Room {
	r := Room{
		name:     roomname,
		users:    []*User{},    //starts empty
		messages: []*Message{}, //starts empty
	}
	roomLock.Lock()
	if val, ok := rooms[roomname]; ok == true {
		log.Debug("Duplicate room. ")
		return val
	}
	rooms[roomname] = &r
	roomLock.Unlock()
	return &r
}

// RemoveRoom deletes a Room from the Rooms list
func RemoveRoom(roomname string) {
	roomLock.Lock()
	_, ok := rooms[roomname]
	if ok {
		delete(rooms, roomname)
	}
	roomLock.Unlock()
}

// CreateUser creates a new user with name username and connection pointer c
func CreateUser(username string, c *net.Conn) *User {
	u := User{
		name:    username,
		conn:    c,
		channel: make(chan *Message),
	}
	userLock.Lock()
	users[username] = &u
	userLock.Unlock()
	return &u
}

// RemoveUser removes a user from the Users list
func RemoveUser(username string) {
	userLock.Lock()
	_, ok := users[username]
	if ok {
		delete(users, username)
	}
	userLock.Unlock()
}

// Start runs the server. Ta-dah.
func Start() {
	setupLogging()
	log.Notice("Starting cryptchat server...")
	Listen(ReadHandler)
}

func timeResponse() []byte {
	t := time.Now()
	return []byte("[" + t.Format("3:04PM") + "]:")
}

func writePrompt(c *net.Conn) {
	(*c).Write(timeResponse())
}

func readInput(c *net.Conn) (int, []byte, error) {
	buff := make([]byte, 1024*4) // This will have to be fiddled with
	count, err := (*c).Read(buff)
	if err != nil {
		// handle, log, etc...
		return count, nil, err
	}
	return count, buff, nil
}

// got this setup from the github page for go-logging
// makes a nice clean colored logging output.
func setupLogging() {
	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(os.Stderr, "", 0)

	// For messages written to backend2 we want to add some additional
	// information to the output, including the used log level and the name of
	// the function.
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")

	// Set the backends to be used.
	logging.SetBackend(backend1Leveled, backend2Formatter)

}
