package clientroute

import (
	"net/rpc"
	"bytes"
)

type ClientRoute struct{
	client *rpc.Client
	key string

	buf *bytes.Buffer
}

func NewClientRoute (client *rpc.Client, pattern string) (route *ClientRoute, err error) {
	route = new(ClientRoute)
	err = client.Call(`Routing.NewRoute`, pattern, &route.key)

	if err != nil {
		return nil, err
	}

	route.client = client	
	route.buf = bytes.NewBuffer(nil)

	return route, nil
}

func (c *ClientRoute) Read (p []byte) (n int, err error) {
	if c.buf.Len() == 0 {
		var buf []byte
		err = c.client.Call("Routing.ReadFrom", c.key, &buf)
		
		if err != nil {
			return
		}

		c.buf = bytes.NewBuffer(buf)
	}

	n, err = c.buf.Read(p)
	return
}

func (c *ClientRoute) Write (p []byte) (n int, err error) {
	err = c.client.Call("Remote.Send", p, &n)
	return
}

func (c *ClientRoute) Close () (err error) {
	return c.client.Call("Routing.RemoveRoute", c.key, &err)
}
