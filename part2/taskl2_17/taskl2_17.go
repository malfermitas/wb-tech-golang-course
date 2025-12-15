package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"sync"
	"time"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s [--timeout duration] host port\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	var timeout time.Duration
	flag.DurationVar(&timeout, "timeout", 10*time.Second,
		"connection timeout (e.g. 5s, 500ms)")
	flag.Usage = usage
	flag.Parse()

	args := flag.Args()
	if len(args) != 2 {
		usage()
		os.Exit(2)
	}

	hostname, port := args[0], args[1]

	addr := net.JoinHostPort(hostname, port)

	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	doneChan := make(chan struct{})

	var closeDoneChannelOnce sync.Once
	closeDoneChannelFunc := func() {
		closeDoneChannelOnce.Do(func() {
			close(doneChan)
		})
	}

	var closeConnectionOnce sync.Once
	closeConnectionFunc := func(conn net.Conn) {
		closeConnectionOnce.Do(func() {
			if err := conn.Close(); err != nil {
				fmt.Fprintf(os.Stderr, "Error closing connection: %v\n", err)
			}
			fmt.Println("Connection with", conn.RemoteAddr().String(), "terminated")
		})
	}

	// stdin processing goroutine
	go func() {
		_, err := io.Copy(conn, os.Stdin)
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Fprintf(os.Stderr, "Error sending data to %s: %v\n", addr, err)
		}

		closeConnectionFunc(conn)
		closeDoneChannelFunc()
	}()

	// stdout processing goroutine
	go func() {
		_, err := io.Copy(os.Stdout, conn)
		if err != nil && !errors.Is(err, io.EOF) {
			fmt.Fprintf(os.Stderr, "Error receiving data from %s: %v\n", addr, err)
		}

		closeConnectionFunc(conn)
		closeDoneChannelFunc()
	}()

	<-doneChan
}
