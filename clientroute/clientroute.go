package clientroute

import (
	"io"
	"net/rpc"
)

type ClientRoute struct{
	io.WriteCloser
	client *rpc.Client
	key string
	buf []byte
}

func NewClientRoute (client *rpc.Client, pattern string) (route *ClientRoute, err error) {
	route = new(ClientRoute)
	err = client.Call(`Routing.NewRoute`, pattern, &route.key)

	if err != nil {
		return nil, err
	}
	route.client = client	
	return route, nil
}

//TODO: would be nicer if this satisfied io.Reader, but that isn't going
//      to happen any time soon
func (c *ClientRoute) Recv (p *[]byte) (n int, err error) {
	err = c.client.Call("Routing.RecvFrom", c.key, &p)
	n = len(p)
	return
}

func (c *ClientRoute) Write (p []byte) (n int, err error) {
	err = c.client.Call("Remote.Send", p, &n)
	return
}

func (c *ClientRoute) Close () error {
	err := c.client.Close()
	return err
}
