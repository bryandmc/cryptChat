package cryptchat

import (
	"fmt"
	"net"

	"regexp"

	"time"

	"github.com/mediocregopher/radix.v2/pool"
	"github.com/mediocregopher/radix.v2/redis"
)

// Listen listens for incoming connections then hands it off to another goroutine for handling
// gotta love go ;)
func Listen(handle func(net.Conn)) {
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
				go handle(c) // handle that shit in a goroutine so it's not blocking
			}
		}
	}
}
func setupRedisPool() *pool.Pool {
	p, err := pool.New("tcp", "localhost:6379", 10)
	if err != nil {
		// handle
	} else {
		return p
	}
	return nil
}

func postRedis(input string, channel string, pool *pool.Pool) {
	client, err := pool.Get()
	if err != nil {
		// handle
		fmt.Println(err.Error())
	} else {
		defer pool.Put(client)
		c := make([]byte, len(input)+1)
		copy(c[:], input[:])
		cipher, err := Encrypt(c, []byte("Blah blah blah something is a little bit too long")[:32]) // needs to be 32 bytes long
		if err != nil {
			panic(err)
		}
		client.Cmd("PUBLISH", channel, cipher)
		fmt.Println("encoded:", string(cipher))
		dec, err := Decrypt(cipher, []byte("Blah blah blah something is a little bit too long")[:32])
		if err != nil {
			panic(err) // TODO: handle errors the same way
		}
		fmt.Println("*************************************************")
		fmt.Println("decode:", string(dec))
	}
}

func ListenRedis(c *net.Conn, rconn *redis.Client) {

}

// Start runs the server. Ta-dah.
func Start() {
	pool := setupRedisPool()

	//handler for the connections...
	handler := func(c net.Conn) {
		fmt.Println("connected...")
		for {
			writePrompt(&c)
			count, buff, _ := readInput(&c)
			if count <= 0 {
				return // get out of here
			}
			//userInput, err := ioutil.ReadAll(c)
			s := string(buff[:count])
			re := regexp.MustCompile("\n")
			index := re.FindStringIndex(s) // gives it in [start, end] []int format
			fmt.Println("Index positions: ", index)
			go postRedis(s[0:index[0]], "channel", pool)
			defer c.Close()
		}
	}
	Listen(handler)
}

func timeResp() []byte {
	t := time.Now()
	return []byte("[" + t.Format("3:04PM") + "]:")
}

func writePrompt(c *net.Conn) {
	(*c).Write(timeResp())
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
