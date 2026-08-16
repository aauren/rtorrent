// Package rtorrent implements a client for rTorrent.
package rtorrent

import (
	"fmt"
	"net/http"

	"github.com/kolo/xmlrpc"
)

//go:generate go tool mockgen -source=rtorrent.go -destination=rtorrent_moq.go -package=rtorrent -typed

type Client interface {
	Close() error
	DownloadTotal() (int, error)
	UploadTotal() (int, error)
	DownloadRate() (int, error)
	UploadRate() (int, error)

	getSliceSlice(method string, args ...string) ([][]any, error)
	getSliceSliceByHash(method string, args ...string) ([][]any, error)
	getStringSlice(method string, args ...string) ([]string, error)
	getInt(method string, arg string) (int, error)
	getString(method string, arg string) (string, error)
}

// A XMLRPCClient is an rTorrent client.  It can be used to retrieve a variety of statistics from rTorrent.
type XMLRPCClient struct {
	xrc *xmlrpc.Client
}

// New creates a new Client using the input XML-RPC address and an optional transport.  If transport is nil, a default one will be used.
func New(addr string, transport http.RoundTripper) (Client, error) {
	xrc, err := xmlrpc.NewClient(addr, transport)
	if err != nil {
		return nil, fmt.Errorf("creating xml-rpc client for %q: %w", addr, err)
	}

	c := &XMLRPCClient{
		xrc: xrc,
	}

	return c, nil
}

// Close frees a Client's resources.
func (c *XMLRPCClient) Close() error {
	return c.xrc.Close()
}

// DownloadTotal retrieves the total number of downloaded bytes since rTorrent startup.
func (c *XMLRPCClient) DownloadTotal() (int, error) {
	return c.getInt("down.total", "")
}

// UploadTotal retrieves the total number of uploaded bytes since rTorrent startup.
func (c *XMLRPCClient) UploadTotal() (int, error) {
	return c.getInt("up.total", "")
}

// DownloadRate retrieves the current download rate in bytes from rTorrent.
func (c *XMLRPCClient) DownloadRate() (int, error) {
	return c.getInt("down.rate", "")
}

// UploadRate retrieves the current upload rate in bytes from rTorrent.
func (c *XMLRPCClient) UploadRate() (int, error) {
	return c.getInt("up.rate", "")
}

// call runs the XML-RPC method and decodes into out, tagging failures with the method name because transport errors
// on their own give no clue as to which call went wrong
func (c *XMLRPCClient) call(method string, send any, out any) error {
	if err := c.xrc.Call(method, send, out); err != nil {
		return fmt.Errorf("xml-rpc call %q: %w", method, err)
	}
	return nil
}

// argsToAny widens the string args into the []any the XML-RPC codec expects, prefixed by the lead arguments
func argsToAny(lead []any, args []string) []any {
	send := make([]any, 0, len(lead)+len(args))
	send = append(send, lead...)
	for _, a := range args {
		send = append(send, a)
	}
	return send
}

// getInt retrieves an integer value from the specified XML-RPC method.
func (c *XMLRPCClient) getInt(method string, arg string) (int, error) {
	var send any
	if arg != "" {
		send = arg
	}

	var v int
	return v, c.call(method, send, &v)
}

// getString retrieves a string value from the specified XML-RPC method.
func (c *XMLRPCClient) getString(method string, arg string) (string, error) {
	var send any
	if arg != "" {
		send = arg
	}

	var v string
	return v, c.call(method, send, &v)
}

// getStringSlice retrieves a slice of string values from the specified XML-RPC method.
func (c *XMLRPCClient) getStringSlice(method string, args ...string) ([]string, error) {
	var v []string
	return v, c.call(method, argsToAny([]any{""}, args), &v)
}

// getSliceSlice retrieves a slice of slice values from the specified XML-RPC method.
func (c *XMLRPCClient) getSliceSlice(method string, args ...string) ([][]any, error) {
	var v [][]any
	return v, c.call(method, argsToAny([]any{""}, args), &v)
}

// getSliceSliceByHash retrieves a slice of slice values scoped to the info-hash that must be passed as the first argument.
func (c *XMLRPCClient) getSliceSliceByHash(method string, args ...string) ([][]any, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("%w: %s requires an info-hash as its first argument", ErrBadData, method)
	}

	var v [][]any
	return v, c.call(method, argsToAny([]any{args[0], ""}, args[1:]), &v)
}
