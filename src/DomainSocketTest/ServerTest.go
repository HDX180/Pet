package main

import (
	"DomainSocket"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
)

func client(u *net.UnixConn, i int) {
	if 1 == i {
		fmt.Println("A client connected : " + u.RemoteAddr().String())
	} else {
		fmt.Println("A client disconnected : " + u.RemoteAddr().String())
	}
}

func reciveMsg(u *net.UnixConn, s string) {
	fmt.Println("A client: " + u.RemoteAddr().String() + " send" + s)
}

func main() {
	s := DomainSocket.NewServer()
	s.Listen("/tmp/unix_sock", client, reciveMsg, '\n')

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Kill, os.Interrupt, syscall.SIGTERM)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.", sig)
		s.Stop()
		os.Exit(0)
	}(sigc)

	defer s.Stop()

	sig := make(chan int, 1)

	go func(c chan int) {
		var input string
		for {
			fmt.Print("输入stop结束")
			fmt.Scanln(&input)
			if input == "stop" {
				break
			}
			c <- 1
		}
	}(sig)

	i := <-sig
	log.Printf("Caught sig %d: stop.", i)

}
