package main

import (
	"flag"
	"fmt"

	"log"
	"net"
	"os"

	//"github.com/hashicorp/yamux"
	"./yamux"
	"time"
)

var session *yamux.Session
var agentpassword string
var sconn net.Conn

func main() {

	listen := flag.String("listen", "", "pipename")

	socks := flag.String("socks", "", "socks address:port")
	connect := flag.String("connect", "", "connect address:port")

	optpassword := flag.String("pass", "", "Connect password")

	recn := flag.Int("recn", 3, "reconnection limit")

	rect := flag.Int("rect", 30, "reconnection delayin sec")
	version := flag.Bool("version", false, "version information")
	flag.Usage = func() {
		fmt.Println("rsockspipe - reverse socks5 server/client")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("0) Generate self-signed SSL certificate: openssl: openssl req -new -x509 -keyout server.key -out server.crt -days 365 -nodes")
		fmt.Println("1) Start rsockstun -listen :8080 -socks 127.0.0.1:1080 -cert server on the server.")
		fmt.Println("2) Start rsockstun -connect client:8080 on the client inside LAN.")
		fmt.Println("3) Connect to 127.0.0.1:1080 on the server with any socks5 client to access into LAN.")
		fmt.Println("X) Enjoy. :]")
	}

	flag.Parse()

	if *version {
		fmt.Println("rsockspipe - reverse socks5 server/client")
		os.Exit(0)
	}

	if *connect != "" {

		if *optpassword != "" {
			agentpassword = *optpassword
		} else {
			agentpassword = "RocksDefaultRequestRocksDefaultRequestRocksDefaultRequestRocks!!"
		}

		//log.Fatal(connectForSocks(*connect,*proxy))
		if *recn > 0 {
			for i := 1; i <= *recn; i++ {
				log.Printf("Connecting to the far end. Try %d of %d", i, *recn)
				error1 := connectForPipes(*connect, *socks)
				log.Print(error1)
				log.Printf("Sleeping for %d sec...", *rect)
				tsleep := time.Second * time.Duration(*rect)
				time.Sleep(tsleep)
			}

		} else {
			for {
				log.Printf("Reconnecting to the far end... ")
				error1 := connectForPipes(*connect, *socks)
				log.Print(error1)
				log.Printf("Sleeping for %d sec...", *rect)
				tsleep := time.Second * time.Duration(*rect)
				time.Sleep(tsleep)
			}
		}

		log.Fatal("Ending...")
	}

	if *listen != "" {
		log.Println("Starting to listen for clients")

		if *optpassword != "" {
			agentpassword = *optpassword
		} else {
			agentpassword = "RocksDefaultRequestRocksDefaultRequestRocksDefaultRequestRocks!!"
		}

		for {
			err := listenForPipes(*listen, *socks)
			if err != nil {
				log.Printf("Main.go Error %v", err)

			}
		}
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "You must specify a listen port or a connect address")
	os.Exit(1)
}
