package main

import (
	"fmt"
	"github.com/armon/go-socks5"
	"io"
	"log"
	"net"
	"os"

	"bufio"
	"time"
	//"encoding/hex"
	//"github.com/hashicorp/yamux"
	"./yamux"
	"github.com/natefinch/npipe"
	//"strings"
)

//var session *yamux.Session
//var stream *yamux.Stream
var proxytout = time.Millisecond * 2000 //timeout for wait for password
//var rurl string                         //redirect URL

// Catches yamux connecting to us
func listenForPipes(pipename string, socksaddr string) error {
	log.Println("Listening for the far end")
	server, err := socks5.New(&socks5.Config{})

	ln, err := npipe.Listen(`\\.\pipe\` + pipename)

	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()
		//conn.RemoteAddr()
		log.Printf("Got a pipe connection from %v...", conn.RemoteAddr())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Errors accepting!")
			return err
		}

		reader := bufio.NewReader(conn)

		//read only 64 bytes with timeout=1-3 sec. So we haven't delay with browsers
		conn.SetReadDeadline(time.Now().Add(proxytout))
		statusb := make([]byte, 64)
		_, _ = io.ReadFull(reader, statusb)

		//Alternatively  - read all bytes with timeout=1-3 sec. So we have delay with browsers, but get all GET request
		//conn.SetReadDeadline(time.Now().Add(proxytout))
		//statusb,_ := ioutil.ReadAll(magicBuf)

		//log.Printf("magic bytes: %v",statusb[:6])
		//if hex.EncodeToString(statusb) != magicbytes {
		if string(statusb)[:len(agentpassword)] != agentpassword {
			//do HTTP checks
			log.Printf("Received request: %v", string(statusb[:64]))

			conn.Close()

		} else {
			//password is correct

			log.Println("Got remote pipe client")
			conn.SetReadDeadline(time.Now().Add(100 * time.Hour))

			if socksaddr == "" {
				//connect with yamux
				//Add connection to yamux
				yconf := yamux.DefaultConfig()
				yconf.EnableKeepAlive = true
				yconf.KeepAliveInterval = time.Millisecond * 10000
				session, err = yamux.Server(conn, yconf)

				for {
					stream, err := session.Accept()
					log.Println("Acceping yamux stream")
					if err != nil {
						return err
					}
					log.Println("Passing off to socks5")
					go func() {
						err = server.ServeConn(stream)
						if err != nil {
							log.Println(err)
						}
					}()
				}
			} else {
				yconf := yamux.DefaultConfig()
				yconf.EnableKeepAlive = true
				yconf.KeepAliveInterval = time.Millisecond * 50000

				session, err = yamux.Client(conn, yconf)

				if err != nil {
					return err
				}

				ServerListenForClientSocks(socksaddr)

			}

		}
	}
}

// Catches clients and connects to yamux
func ServerListenForClientSocks(address string) error {
	log.Println("Waiting for socks clients")
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		// TODO dial socks5 through yamux and connect to conn

		if session == nil {
			conn.Close()
			continue
		}
		log.Println("Got a socks client")

		log.Println("Opening a stream")
		stream, err := session.Open()

		if err != nil {
			return err
		}

		// connect both of conn and stream

		go func() {
			log.Printf("Starting to copy conn to stream id:  ")

			io.Copy(conn, stream)

			conn.Close()
		}()
		go func() {
			log.Println("Starting to copy stream to conn")

			io.Copy(stream, conn)
			//log.Printf("Closing stream id: %d ",stream)
			stream.Close()

			log.Println("Done copying stream to conn")
		}()
	}
}
