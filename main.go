package main

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
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

func pipe(ctx context.Context, reader, writer net.Conn) func() error {
	return func() error {
		for {
			// checking that sister goroutine is still alive
			select {
			case <-ctx.Done():
				// sister goroutine signals cancel
				log.Println("sister goroutine signals cancel")
				return nil
			default:
				// good to continue
			}
			buffer := make([]byte, 1024)
			n, err := reader.Read(buffer)
			data := buffer[:n]
			if err != nil {
				log.Printf("%v", errors.Wrapf(err, "connection %s", reader.RemoteAddr()))
				return err
			}
			_, err = writer.Write(data)
			if err != nil {
				log.Printf("%v", errors.Wrapf(err, "connection %s", writer.RemoteAddr()))
				return err
			}
		}
	}
}

func handleConnection(readConn net.Conn) {
	defer func() {
		if err := readConn.Close(); err != nil {
			log.Printf("%v", errors.Wrapf(err, "failed closing connection %s", readConn.RemoteAddr()))
		}
	}()
	writeConn, err := net.Dial("tcp", targetAddress)
	if err != nil {
		log.Printf("%+v", errors.Wrapf(err, "failed to create a new connection to target %s", targetAddress))
	}
	defer func() {
		if err := writeConn.Close(); err != nil {
			log.Printf("%v", errors.Wrapf(err, "failed closing connection %s", writeConn.RemoteAddr()))
		}
	}()
	// as this is a bi-directional proxy if one of our connections closes we need to clean up its "sister" connection
	eg, ctx := errgroup.WithContext(context.Background())

	eg.Go(pipe(ctx, readConn, writeConn))
	eg.Go(pipe(ctx, writeConn, readConn))

	_ = eg.Wait() // waiting here to avoid closing the connections prematurely
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
