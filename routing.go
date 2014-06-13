package main

import (
	"net/rpc"

	"regexp"
	"fmt"
	"io"
	"bufio"
)

var (
	routes = make(chan map[string]*route, 1)
)

func init () {
	rpc.Register(new(Routing))

	routes <-make(map[string]*route, 1)
}

type route struct {
	in *bufio.Writer
	out *io.PipeReader

	re *regexp.Regexp
}

func newRoute () *route {
	out, in := io.Pipe()
	return &route{
		out: out, in: bufio.NewWriter(in),
	}
}

func (r *route) Read (p []byte) (n int, err error) {
	n, err = r.out.Read(p)

	return
}

func (r *route) Write (p []byte) (n int, err error) {
	if r.re.Match(p) {
		n, err = r.in.Write(p)
		if err == nil {
			r.in.Flush()
		}
	}
	return
}

type Routing struct{}

func (_ *Routing) NewRoute (pattern string, key *string) error {

	re, err := regexp.Compile(pattern)

	if err != nil {
		return err
	}

	*key = pattern
	route := newRoute()
	route.re = re

	m := <-routes
	m[*key] = route
	routes <-m

	return nil
}

func (_ *Routing) RemoveRoute (k string, err *error) error {
	m := <-routes
	
	m[k].in.Flush()
	m[k].out.Close()
	delete(m, k)

	routes <-m
	return nil
}

func (_ *Routing) ReadFrom (k string, reply *[]byte) error {

	m := <-routes
	route, ok := m[k]
	routes <-m

	if !ok {
		return fmt.Errorf("no such route: %q", k)
	}

	*reply = make([]byte, 1024)
	n, err := route.Read(*reply)

	if err != nil {
		return err
	}

	*reply = (*reply)[:n]

	return nil
}
