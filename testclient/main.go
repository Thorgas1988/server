package main

import (
	"strconv"
	"time"
	"log"
	"os"
	"path/filepath"

	"github.com/jlaffaye/ftp"
)

func main() {
	server := "127.0.0.1"
	port := 2121
	username := "admin"
	password := "123456"
	fileName := "testclient/testfile.xml"

	cli, err := ftp.Dial(server+":"+strconv.Itoa(port), ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		log.Fatal(err)
	}

	err = cli.Login(username, password)
	if err != nil {
		log.Fatal(err)
	}

	fh , err := os.Open(fileName) 
	if err != nil {
		log.Fatal(err)
	}	

	outFile := filepath.Base(fileName)

	err = cli.Stor(outFile, fh)
	if err != nil {
		log.Fatal(err)
	}
	
}
