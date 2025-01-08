// Package rtorrent implements a client for rTorrent.
package rtorrent

import (
	"net/http"

	"github.com/kolo/xmlrpc"
)

// A Client is an rTorrent client.  It can be used to retrieve a variety of statistics from rTorrent.
type Client struct {
	Downloads *DownloadService
	Trackers  *TrackerService

	xrc *xmlrpc.Client
}

// New creates a new Client using the input XML-RPC address and an optional transport.  If transport is nil, a default one will be used.
func New(addr string, transport http.RoundTripper) (*Client, error) {
	xrc, err := xmlrpc.NewClient(addr, transport)
	if err != nil {
		return nil, err
	}

	c := &Client{
		xrc: xrc,
	}

	c.Downloads = &DownloadService{c: c}
	c.Trackers = &TrackerService{C: c}

	return c, nil
}

// Close frees a Client's resources.
func (c *Client) Close() error {
	return c.xrc.Close()
}

// DownloadTotal retrieves the total number of downloaded bytes since rTorrent startup.
func (c *Client) DownloadTotal() (int, error) {
	return c.getInt("down.total", "")
}

// UploadTotal retrieves the total number of uploaded bytes since rTorrent startup.
func (c *Client) UploadTotal() (int, error) {
	return c.getInt("up.total", "")
}

// DownloadRate retrieves the current download rate in bytes from rTorrent.
func (c *Client) DownloadRate() (int, error) {
	return c.getInt("down.rate", "")
}

// UploadRate retrieves the current upload rate in bytes from rTorrent.
func (c *Client) UploadRate() (int, error) {
	return c.getInt("up.rate", "")
}

// getInt retrieves an integer value from the specified XML-RPC method.
func (c *Client) getInt(method string, arg string) (int, error) {
	var send interface{}
	if arg != "" {
		send = arg
	}

	var v int
	err := c.xrc.Call(method, send, &v)
	return v, err
}

// getString retrieves a string value from the specified XML-RPC method.
func (c *Client) getString(method string, arg string) (string, error) {
	var send interface{}
	if arg != "" {
		send = arg
	}

	var v string
	err := c.xrc.Call(method, send, &v)
	return v, err
}

// getStringSlice retrieves a slice of string values from the specified XML-RPC method.
//
//nolint:unparam // we don't care about download_list being the only method so far
func (c *Client) getStringSlice(method string, args ...string) ([]string, error) {
	send := []interface{}{""}
	for _, a := range args {
		send = append(send, a)
	}

	var v []string
	err := c.xrc.Call(method, send, &v)
	return v, err
}

// getSliceSlice retrieves a slice of slice values from the specified XML-RPC method.
func (c *Client) getSliceSlice(method string, args ...string) ([][]any, error) {
	send := []interface{}{""}
	for _, a := range args {
		send = append(send, a)
	}

	var v [][]any
	err := c.xrc.Call(method, send, &v)
	return v, err
}

func (c *Client) getSliceSliceByHash(method string, args ...string) ([][]any, error) {
	send := []interface{}{args[0], ""}
	for _, a := range args[1:] {
		send = append(send, a)
	}

	var v [][]any
	err := c.xrc.Call(method, send, &v)
	return v, err
}
