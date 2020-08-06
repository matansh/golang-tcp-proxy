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

func pipe(reader, writer net.Conn) {
	defer func() { _ = reader.Close() }()
	defer func() { _ = writer.Close() }()

	for {
		buffer := make([]byte, 1024)
		n, err := reader.Read(buffer)
		data := buffer[:n]
		if err != nil {
			if err == io.EOF {
				break // gracefully exiting
			}
			log.Printf("%+v", errors.Wrap(err, "failed reading data"))
			break
		}
		_, err = writer.Write(data)
		if err != nil {
			log.Printf("%+v", errors.Wrap(err, "failed writing data"))
			break
		}
	}
	log.Printf("connection %s ended", reader.RemoteAddr())
}

func handleConnection(readConn net.Conn) {
	writeConn, err := net.Dial("tcp", targetAddress)
	if err != nil {
		log.Printf("%+v", errors.Wrapf(err, "failed to create a new connection to target %s", targetAddress))
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
	ln, err := net.Listen("tcp4", address)
	if err != nil {
		log.Fatalf("%+v", errors.Wrapf(err, "failed to listen on %s", address))
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("%+v", errors.Wrapf(err, "failed to accept a connection on %s", address))
		}
		log.Printf("got a connection from %s", conn.RemoteAddr())
		go handleConnection(conn)
	}

}
