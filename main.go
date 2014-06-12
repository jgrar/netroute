package main

import (
	"flag"
	"log"
	"os"
	"fmt"
	"strings"
	"strconv"

	"net/rpc"
	"net"
	"time"
	"crypto/tls"
	"bufio"
	"io/ioutil"
)

const PROGRAM_NAME = "netroute"

var (
	verbose = flag.Bool("verbose", false, "prints extra program information")
	delim = flag.String("delim", "\n", "specifies delimeter string")
	ssl = flag.Bool("ssl", false, "use ssl to connect to remote host")
	listenport = flag.String("port", "6666", "specify listen port")

	host, port string

	usage = func () {
		fmt.Fprintf(os.Stderr, "usage: %s host port\n", os.Args[0])
		flag.PrintDefaults()
	}

	ERROR = log.New(os.Stderr, PROGRAM_NAME + ": ERROR: ", log.LstdFlags)


	sslcfg = tls.Config{
		InsecureSkipVerify: true,
	}

	remote net.Conn
)


func init () {
	flag.Usage = usage

	rpc.Register(new(Remote))
}

func main () {
	flag.Parse()

	if flag.NArg() < 2 {
		ERROR.Println("missing arguments")
		flag.Usage()
		os.Exit(1)
	}

	if !*verbose {
		log.SetOutput(ioutil.Discard)
	}

	host, port = flag.Arg(0), flag.Arg(1)

	if err := RunNetRoute(); err != nil {
		ERROR.Fatal(err)
	}
}

func RunNetRoute () error {


	listen, err := net.Listen("tcp", ":" + *listenport)
	if err != nil {
		return err
	}

	con, err := listen.Accept()
	if err != nil {
		listen.Close()
		return err

	}

	remote, err = net.DialTimeout("tcp", host + ":" + port, 30 * time.Second)
	if err != nil {
		listen.Close()
		return err
	}

	if *ssl {
		sslcfg.ServerName = host
		remote = tls.Client(remote, &sslcfg)
	}

	go rpc.ServeConn(con)

	shutdown := make(chan error, 1)

	go func () {
		s := bufio.NewScanner(remote)

		sf, err := scanDelim(*delim)
		if err != nil {
			shutdown <-err
		}
		s.Split(sf)

		for s.Scan() {
			msg := make([]byte, len(s.Bytes()))
			copy(msg, s.Bytes())

			log.Printf(">> %s", string(msg))

			m := <-routes
			for _, f := range m {
				f.Write(msg)
			}
			routes <-m
		}

		shutdown <-s.Err()
	}()

	go func () {
		for {
			con, err := listen.Accept()

			if err != nil {
				shutdown <-err
				return
			}

			go rpc.ServeConn(con)
		}
	}()

	err = <-shutdown

	listen.Close()
	remote.Close()

	return err
}

func scanDelim (delim_ string) (bufio.SplitFunc, error) {

	delim, err := unquote(delim_)
	if err != nil {
		return nil, err
	}

	return func (data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
	
		if i := strings.Index(string(data), delim); i >= 0 {
			return i + len(delim), data[:i], nil
		}

		if atEOF {
			return len(data), data, nil
		}

		return 0, nil, nil
	}, nil

}

// TODO: this is here because strconv.Unquote("\\r\\n") doesn't do
//       what it is supposed to, there has to be a better way
func unquote (in string) (string, error) {
	var (
		c rune
		out []rune
		err error
	)

	for len(in) > 0 {
		c, _, in, err = strconv.UnquoteChar(in, '"')
		if err != nil {
			break
		}
		
		out = append(out, c)
	}
	return string(out), err
}

type Remote struct {}

func (_ *Remote) Send (p []byte, n *int) (err error) {
	*n, err = remote.Write(p)
	if err == nil {
		log.Printf("<< %s\n", string(p))
	}
	return
}

