package main

import (
	"github.com/armon/go-socks5"
	"io"
	"log"
	"net"

	//"github.com/hashicorp/yamux"
	"./yamux"
	"github.com/natefinch/npipe"
	"strings"
	"time"
)

//var encBase64 = base64.StdEncoding.EncodeToString
//var decBase64 = base64.StdEncoding.DecodeString
//var password string
//var proxytimeout = time.Millisecond * 1000 //timeout for proxyserver response

func connectForPipes(pipeaddr string, socksaddr string) error {

	server, err := socks5.New(&socks5.Config{})
	if err != nil {
		return err
	}

	//

	log.Println("Connecting to far end with 5 sec timeout")
	serveraddr := strings.Split(pipeaddr, "\\")[0]
	pipename := strings.Split(pipeaddr, "\\")[1]

	conn, err := npipe.DialTimeout(`\\`+serveraddr+`\pipe\`+pipename, time.Second*5)
	if err != nil {
		return err
	}

	log.Println("Starting client")

	conn.Write([]byte(agentpassword))
	//time.Sleep(time.Second * 1)

	if socksaddr != "" {
		yconf := yamux.DefaultConfig()
		yconf.EnableKeepAlive = true
		yconf.KeepAliveInterval = time.Millisecond * 50000

		session, err = yamux.Client(conn, yconf)

		if err != nil {
			return err
		}

		ClientListenForClientSocks(socksaddr)
	} else {
		yconf := yamux.DefaultConfig()
		yconf.EnableKeepAlive = false
		yconf.KeepAliveInterval = time.Millisecond * 50000
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

	}

	return nil
}

// Catches clients and connects to yamux
func ClientListenForClientSocks(address string) error {
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
