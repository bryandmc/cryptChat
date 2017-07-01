package cryptchat

import (
	"fmt"
	"net"
	"os"

	"regexp"

	"time"

	"sync"

	logging "github.com/op/go-logging"
)

// Users contains a list of the currently connected User objects
var Users = make(map[string]*User)
var userLock = &sync.RWMutex{}

// Rooms contains a list of the currently available/active Room objects
var Rooms = make(map[string]*Room)
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
		val := fmt.Errorf("Error: %s", err.Error())
		fmt.Println(val)
	} else {
		for { // keeps accepting connections forever
			c, err := conn.Accept()
			if err != nil {
				// handle
			} else { // only want to handle if there WASN'T an error
				go handle(&c) // handle that shit in a goroutine so it's not blocking
			}
		}
	}
}

// RecieveMsgs blocks until it gets a message, sends that the the user and
// repeats. Nice and simple.
func RecieveMsgs(usr *User) {
	for {
		val := <-usr.channel
		c := *usr.conn
		c.Write([]byte("[" + val.sentFrom.name + "] " + val.msg + "\n"))
		//log.Error(<-usr.channel)
	}
}

// ReadHandler handles all the incoming connections and reading from the
// socket. Little more complicated than I'd like currently.
var ReadHandler = func(c *net.Conn) {
	usr := CreateUser("Bryan", c)
	go RecieveMsgs(&usr)
	log.Debug(usr.name)
	//log.Info(Users) //unsafe!
	rm := CreateRoom("test")
	JoinRoom(&usr, &rm)
	log.Debug(rm)
	log.Debug("Starting goroutine to handle connection from:", (*c).RemoteAddr())
	for {
		//writePrompt(c)
		count, buff, _ := readInput(c)
		if count <= 0 {
			log.Warningf("No data read (%d) closing goroutine.", count)
			return // get out of here
		}
		s := string(buff[:count]) // convert to string regex
		re := regexp.MustCompile("\n")
		index := re.FindStringIndex(s) // gives it in [start, end] []int format
		log.Info("Index position of newline character:", index)

		// send messages here
		var msg = Message{
			msg:        s[:index[0]],
			attachment: []byte{},
			sentFrom:   &usr,
			sentTo:     &usr,
			room:       nil,
			isToRoom:   false,
		}
		msg.Send()
		msg.Send()
		//log.Critical(s[:index[0]])
	}
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
func CreateRoom(roomname string) Room {
	r := Room{
		name:     roomname,
		users:    []*User{},    //starts empty
		messages: []*Message{}, //starts empty
	}
	roomLock.Lock()
	Rooms[roomname] = &r
	roomLock.Unlock()
	return r
}

// RemoveRoom deletes a Room from the Rooms list
func RemoveRoom(roomname string) {
	roomLock.Lock()
	_, ok := Rooms[roomname]
	if ok {
		delete(Rooms, roomname)
	}
	roomLock.Unlock()
}

// CreateUser creates a new user with name username and connection pointer c
func CreateUser(username string, c *net.Conn) User {
	u := User{
		name:    username,
		conn:    c,
		channel: make(chan *Message),
	}
	userLock.Lock()
	Users[username] = &u
	userLock.Unlock()
	return u
}

// RemoveUser removes a user from the Users list
func RemoveUser(username string) {
	userLock.Lock()
	_, ok := Users[username]
	if ok {
		delete(Users, username)
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
