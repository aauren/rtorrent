// Package rtorrent implements a client for rTorrent.
package rtorrent

import (
	"fmt"
	"net/http"
	"strings"

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
	commandByHash(method string, args ...string) error
	multicallByHash(infoHash string, methods ...string) ([]any, error)
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

// commandByHash runs a command whose target is the info-hash passed as the first argument, discarding the result
func (c *XMLRPCClient) commandByHash(method string, args ...string) error {
	if len(args) == 0 || args[0] == "" {
		return fmt.Errorf("%w: %s requires an info-hash as its first argument", ErrBadData, method)
	}

	var v any
	return c.call(method, argsToAny(nil, args), &v)
}

// multicallByHash runs each method against infoHash in a single system.multicall, returning one value per method.
// A fault on any one method fails the whole call, carrying the xmlrpc.FaultError rTorrent sent for it.
func (c *XMLRPCClient) multicallByHash(infoHash string, methods ...string) ([]any, error) {
	if infoHash == "" {
		return nil, fmt.Errorf("%w: system.multicall requires an info-hash", ErrBadData)
	}

	calls := make([]multicallEntry, 0, len(methods))
	for _, m := range methods {
		calls = append(calls, multicallEntryFor(infoHash, m))
	}

	var raw []any
	if err := c.call("system.multicall", []any{calls}, &raw); err != nil {
		return nil, err
	}
	if len(raw) != len(methods) {
		return nil, fmt.Errorf("%w: system.multicall returned %d results for %d methods", ErrBadData, len(raw), len(methods))
	}

	out := make([]any, 0, len(raw))
	for _, r := range raw {
		v, err := multicallValue(r)
		if err != nil {
			return nil, fmt.Errorf("xml-rpc call %q: %w", methods[len(out)], err)
		}
		out = append(out, v)
	}
	return out, nil
}

// multicallEntry is a struct rather than a map because newer rTorrent parsers need methodName before params, and
// map iteration order would shuffle them
type multicallEntry struct {
	MethodName string `xmlrpc:"methodName"`
	Params     []any  `xmlrpc:"params"`
}

// multicallEntryFor turns a d.multicall2-style command into a system.multicall entry so the same "d.custom=label"
// spelling works in both places. rTorrent splits the part after "=" on commas, so we do the same.
func multicallEntryFor(infoHash, command string) multicallEntry {
	name, rest, hasArgs := strings.Cut(command, "=")
	params := []any{infoHash}
	if hasArgs && rest != "" {
		for a := range strings.SplitSeq(rest, ",") {
			params = append(params, a)
		}
	}
	return multicallEntry{MethodName: name, Params: params}
}

// multicallValue unwraps one system.multicall result, which is a one-element array on success or a fault struct
func multicallValue(r any) (any, error) {
	switch v := r.(type) {
	case []any:
		if len(v) != 1 {
			return nil, fmt.Errorf("%w: expected 1 value, got %d", ErrBadData, len(v))
		}
		return v[0], nil
	case map[string]any:
		// Only a struct carrying faultCode is a fault, anything else is a result shape we don't know how to unwrap
		if _, ok := v["faultCode"]; !ok {
			return nil, fmt.Errorf("%w: unexpected multicall result %T", ErrBadData, r)
		}
		fault := xmlrpc.FaultError{}
		fault.Code, _ = intFromAny(v["faultCode"])
		fault.String, _ = stringFromAny(v["faultString"])
		return nil, fault
	default:
		return nil, fmt.Errorf("%w: unexpected multicall result %T", ErrBadData, r)
	}
}
