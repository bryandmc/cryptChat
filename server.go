package cryptchat

import (
	"encoding/json"
	"net"
	"os"

	"sync"

	"strings"

	"errors"

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
			//log.Notice([]byte(val.Body))
			c.Write([]byte(val.Body))
		case <-(*quit):
			log.Debug("Quitting goroutine!")
			return
		}
	}
}

// ReadHandler handles all the incoming connections and reading from the
// socket. Little more complicated than I'd like currently.
var ReadHandler = func(c *net.Conn) {
	quit := make(chan bool, 1)
	// local variables related to just this users connection..?
	cmdChan := make(chan *Command, 20)
	go HandleCommand(cmdChan, quit, c) // just get rid of it and let this go back to reading input
	// main read loop
	log.Debug("Starting goroutine to handle connection from:", (*c).RemoteAddr())
	for {
		count, buff, _ := readInput(c)
		if count <= 0 {
			log.Warningf("No data read (%d) closing goroutine.", count)
			quit <- true
			close(quit)
			return // get out of here
		}
		var cmd = Command{}
		err := json.Unmarshal(buff[:count], &cmd)
		if err != nil {
			log.Error("Unmarshal: ", err.Error())

		} else {
			log.Critical("before send cmdChan")
			log.Notice(cmd.Msg.Body)
			cmdChan <- &cmd
		}
	}
}

func HandleCommand(cmdChan chan *Command, quit chan bool, c *net.Conn) {
	for {
		log.Critical("handlecommand")
		cmd := <-cmdChan
		switch cmd.Cmd {
		case SEND_USERNAME:
			log.Debug("send_username")
			username := cmd.Args["connect_username"]
			usr := CreateUser(username, c)
			go RecieveMsgs(usr, &quit)
			log.Debug("USERS (send_username):", users)
		case SEND_DIRECT:
			SendDirect(cmd)
		case SEND_ROOM:
			log.Debug("Send_room")
			//handle
		case JOIN_ROOM:
			log.Debug("join_room")
		case CREATE_ROOM:
			log.Debug("create_room")
		case REMOVE_ROOM:
			log.Debug("remove_room")
		case QUIT:
			log.Debug("quit")
		}
	}
}
func SendDirect(cmd *Command) error {
	log.Debug("Send_Direct")
	to := LookupUser(cmd.Args["to_username"])
	from := LookupUser(cmd.Args["from_username"])
	cmd.Msg.SentTo = to
	cmd.Msg.SentFrom = from
	if to != nil && from != nil {
		cmd.Msg.Send()
		return nil
	}
	return errors.New("could not determind user to/from properly")
}

func LookupUser(username string) *User {
	log.Debug("LookupUser")
	username = strings.Trim(username, "\n")
	userLock.Lock()
	defer userLock.Unlock()
	log.Debug(users[username])
	return users[strings.TrimSpace(username)]
}

// JoinRoom adds a user to a room.
func JoinRoom(usr *User, rm *Room) {
	roomLock.Lock()
	for _, val := range rm.Users { // check for existing user
		if usr == val {
			log.Info(val)
			return
		}
	}
	rm.Users = append(rm.Users, usr)
	roomLock.Unlock()
}

// CreateRoom handles creating a new Room to be inserted into Rooms
func CreateRoom(roomname string) *Room {
	r := Room{
		Name:     roomname,
		Users:    []*User{},    //starts empty
		Messages: []*Message{}, //starts empty
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
		Name:    username,
		conn:    c,
		channel: make(chan *Message),
	}
	userLock.Lock()
	defer userLock.Unlock()
	users[username] = &u
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

func writePrompt(c *net.Conn) {
	(*c).Write(TimeResponse())
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
func main() {
	Start()
}
