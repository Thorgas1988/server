package main

import (
	"context"	
	"net"
	"log"
	"time"
	"github.com/tevino/tcp-shaker"
)

func dialCheck(target string) {	
	conn, err := net.Dial("tcp", target)
	if err != nil {
		log.Fatal("no connection")
	}

	log.Println("Dial Check success, remote port is open", target, conn.RemoteAddr())

	err = conn.Close()
	if err != nil {
		log.Println("Unable to close connection")
	}

	// ok now lets no do anything not writing not readin ...
	// fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	// status, err := bufio.NewReader(conn).ReadString('\n')
	// log.Println(status)

	// for {}
}

func dialCheckHanging(target string) {	
	conn, err := net.Dial("tcp", target)
	if err != nil {
		log.Fatal("no connection")
	}

	log.Println("Dial Check success, remote port is open", target, conn.RemoteAddr())

	log.Println("sleeping")
	time.Sleep(10 * time.Second)
	log.Println("sleeping done")

	// ok now lets no do anything not writing not readin ...
	// fmt.Fprintf(conn, "GET / HTTP/1.0\r\n\r\n")
	// status, err := bufio.NewReader(conn).ReadString('\n')
	// log.Println(status)
	// for {}
}

func tcpCheck(target string) {
	c := tcp.NewChecker()

	ctx, stopChecker := context.WithCancel(context.Background())
	defer stopChecker()
	go func() {
		if err := c.CheckingLoop(ctx); err != nil {
			log.Println("checking loop stopped due to fatal error: ", err)
		}
	}()

	<-c.WaitReady()

	timeout := time.Second * 1
	err := c.CheckAddr(target, timeout)
	switch err {
	case tcp.ErrTimeout:
		log.Println("Connect to ftp timed out")
	case nil:
		log.Println("Tcp Two Way Connect to ftp succeeded")
	default:
		log.Println("Error occurred while connecting: ", err)
	}
}

func main() {	
	dialCheck("127.0.0.1:2121")
	dialCheckHanging("127.0.0.1:2121")
	tcpCheck("127.0.0.1:2121")	
}
