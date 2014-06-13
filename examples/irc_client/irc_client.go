package main

import (
	"net/rpc"
	"log"
	cr "github.com/jgrar/netroute/clientroute"
)

func main () {
	log.SetFlags(log.Lshortfile)

	con, err := rpc.Dial("tcp", ":6666")
	if err != nil {
		log.Fatal(err)
	}

	var n int

	err = con.Call("Remote.Send",
		[]byte("NICK netrouteclient\r\n" +
			"USER user 0 * :real\r\n"), &n)

	go func () {
		reply := make([]byte, 1024)

		c, err := cr.NewClientRoute(con, ":[^ ]+ 001 ")
		if err != nil {
			log.Fatal(err)
		}

		_, err = c.Read(reply)
		if err != nil {
			log.Fatal(err)
		}

		_, err = c.Write([]byte("JOIN #netroute\r\n"))
		if err != nil {
			log.Fatal(err)
		}

		c.Close()
	} ()

	c, err := cr.NewClientRoute(con, "^PING")
	if err != nil {
		log.Fatal(err)
	}

	reply := make([]byte, 1024)

	for {
		_, err = c.Read(reply)
		if err != nil {
			log.Fatal(err)
		}

		log.Println(string(reply))
	
		reply[1] = 'O'

		_, err = c.Write(append(reply, '\r', '\n'))
		if err != nil {
			log.Fatal(err)
		}
	}
}
