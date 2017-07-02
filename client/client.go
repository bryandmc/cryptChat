package main

import (
	"fmt"
	"net"
	"time"

	"bufio"
	"os"

	"github.com/fatih/color"
)

func printBanner() {
	fmt.Println()
	color.Blue("\t------------------------------------------")
	color.Yellow("\t\tWelcome to cryptchat.")
	color.Blue("\t------------------------------------------")
	fmt.Println()
	fmt.Println()
}

func Connect(host string) {
	printBanner()
	t, err := time.ParseDuration("10s")
	if err != nil {
		fmt.Println(err.Error())
	}
	conn, err := net.DialTimeout("tcp", host, t)
	for {
		if err != nil {
			fmt.Println(err.Error())
		}
		b := make([]byte, 1024*10)
		conn.Read(b)
		col := color.New(color.FgBlack).Add(color.BgCyan)
		col.Print(string(b))
		reader := bufio.NewReader(os.Stdin)
		text, _ := reader.ReadString('\n')
		conn.Write([]byte(text))
	}
}
