package main

import (
	"fmt"
	"github.com/pkg/errors"
	"io"
	"log"
	"net"
	"os"
)

// env vars
var host string
var port string
var targetHost string
var targetPort string

// non env global vars
var targetAddress string

func pipe(reader io.Reader, writer io.Writer) {
	for {
		buffer := make([]byte, 1024)
		n, err := reader.Read(buffer)
		data := buffer[:n]
		if err != nil {
			if err == io.EOF {
				break // gracefully exiting
			}
			log.Fatalf("%+v", errors.Wrap(err, "failed reading data"))
		}
		_, err = writer.Write(data)
		if err != nil {
			log.Fatalf("%+v", errors.Wrap(err, "failed writing data"))
		}
	}
}

func handleConnection(readConn net.Conn) {
	target, err := net.ResolveTCPAddr("tcp", targetAddress)
	if err != nil {
		log.Fatalf("%+v", errors.Wrapf(err, "failed resolving target %s", targetAddress))
	}
	writeConn, err := net.DialTCP("tcp", nil, target)
	if err != nil {
		log.Fatalf("%+v", errors.Wrapf(err, "failed to create a new connection to target %s", targetAddress))
	}
	go pipe(readConn, writeConn)
	go pipe(writeConn, readConn)
}

func main() {
	// reading env vars
	host = os.Getenv("HOST")
	port = os.Getenv("PORT")
	targetHost = os.Getenv("TARGET_HOST")
	targetPort = os.Getenv("TARGET_PORT")
	targetAddress = fmt.Sprintf("%s:%s", targetHost, targetPort)

	address := fmt.Sprintf("%s:%s", host, port)
	log.Printf("serving tcp proxy on %s", address)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("%+v", errors.Wrapf(err, "failed to listen on %s", address))
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalf("%+v", errors.Wrapf(err, "failed to accept a connection on %s", address))
		}
		log.Printf("got a connection from %s", conn.RemoteAddr())
		go handleConnection(conn)
	}

}
