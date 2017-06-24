package cryptChat

import (
	"fmt"
	"net"

	"github.com/mediocregopher/radix.v2/pool"
)

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

func postRedis(input string, pool *pool.Pool) {
	client, err := pool.Get()
	if err != nil {
		// handle
		fmt.Println(err.Error())
	} else {
		client.Cmd("PUBLISH", "channel", input)
	}
	fmt.Print(input)
}

func Start() {
	pool := setupRedisPool()
	handler := func(c net.Conn) {
		fmt.Println("connected...")
		buff := make([]byte, 1024)
		for {
			count, err := c.Read(buff)
			if err != nil {
				// handle error
			} else {
				s := string(buff[:count])
				//fmt.Println(s) // for testing
				go postRedis(s, pool)
			}
		}
	}
	Listen(handler)
}
