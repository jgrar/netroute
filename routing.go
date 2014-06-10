package main

import (
	"net/rpc"

	"regexp"
	"fmt"
	"sync"
	"io"
)

var (
	routes = make(chan map[string]*route, 1)
)

func init () {
	rpc.Register(new(Routing))

	routes <-make(map[string]*route, 1)
}

type route  struct {
	sync.RWMutex
	recv chan []byte
	re *regexp.Regexp
}

func newRoute () *route {
	return &route{
		recv: make(chan []byte, 1024),
	}
}

func (r *route) Recv () (msg []byte, err error) {

	r.RLock()
	defer func () {r.RUnlock()} ()

	if len(r.recv) == 0 {
		r.RUnlock()
		msg = <-r.recv
		r.RLock()
	} else {
		msg = <-r.recv
	}

	if msg == nil {
		err = io.EOF
	}
	return
}

func (r *route) Write (p []byte) (n int, err error) {
	if r.re.Match(p) {

		r.Lock()
		r.recv <-p
		r.Unlock()

		n = len(p)
	}
	return n, nil
}

func (r *route) Close () error {
	close(r.recv)
	return nil
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
	
	m[k].Close()
	delete(m, k)

	routes <-m
	return nil
}

func (_ *Routing) RecvFrom (k string, reply *[]byte) (err error) {

	m := <-routes
	route, ok := m[k]
	routes <-m

	if ok {
		*reply, err = route.Recv()
	} else {
		err = fmt.Errorf("route %q does not exist", k)
	}
	return err
}

