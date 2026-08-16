package rtorrent

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testBytes = 1024

func TestClientTotalsAndRates(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		call   func(Client) (int, error)
	}{
		{"download total", "down.total", Client.DownloadTotal},
		{"upload total", "up.total", Client.UploadTotal},
		{"download rate", "down.rate", Client.DownloadRate},
		{"upload rate", "up.rate", Client.UploadRate},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			c := testClient(t, tt.method, nil, testBytes)

			got, err := tt.call(c)
			require.NoError(t, err)
			assert.Equal(t, testBytes, got)
		})
	}
}

func TestGetSliceSliceByHashRequiresInfoHash(t *testing.T) {
	t.Parallel()

	c := &XMLRPCClient{}
	_, err := c.getSliceSliceByHash(trackerListMultiCall)
	require.ErrorIs(t, err, ErrBadData)
}

// testClient stands up an XML-RPC server that asserts the request matches method and wantParams, then replies with out.
// The returned Client is closed along with the server when the test finishes.
func testClient(t *testing.T, method string, wantParams []string, out any) Client {
	t.Helper()

	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// IMPORTANT: this runs on the server's goroutine, so t.Fatalf would Goexit the handler rather than the
		// test, hanging the client on a torn-off response. Only t.Errorf is safe here.
		var xr xmlrpcRequest
		if err := xml.NewDecoder(r.Body).Decode(&xr); err != nil {
			t.Errorf("failed to decode XML-RPC body: %v", err)
			return
		}

		if xr.MethodName != method {
			t.Errorf("unexpected XML-RPC method name:\n- want: %q\n-  got: %q", method, xr.MethodName)
			return
		}

		params := make([]string, 0, len(xr.Params.Param))
		for _, p := range xr.Params.Param {
			params = append(params, p.Value.String)
		}

		if !slices.Equal(wantParams, params) {
			t.Errorf("unexpected XML-RPC parameters:\n- want: %v\n-  got: %v", wantParams, params)
			return
		}

		if err := writeXMLRPC(w, out); err != nil {
			t.Errorf("unexpected error encoding XML-RPC response: %v", err)
		}
	}))

	c, err := New(s.URL, nil)
	require.NoError(t, err, "failed to create Client")

	t.Cleanup(func() {
		assert.NoError(t, c.Close(), "failed to clean up Client")
		s.Close()
	})

	return c
}

// XML-RPC helper routines and structures

func writeXMLRPC(w io.Writer, out any) error {
	var value xmlrpcValue

	switch out := out.(type) {
	case int:
		value.Int = out
	case string:
		value.String = out
	case []string:
		value.Array = new(xmlrpcArray)
		value.Array.Data.Value = make([]xmlrpcArrayData, len(out))

		for i, s := range out {
			value.Array.Data.Value[i].String = s
		}
	}

	var xw xmlrpcResponse
	xw.Params.Param = []xmlrpcParam{{Value: value}}

	return xml.NewEncoder(w).Encode(xw)
}

type xmlrpcRequest struct {
	XMLName    xml.Name `xml:"methodCall"`
	MethodName string   `xml:"methodName"`

	Params xmlrpcParams `xml:"params"`
}

type xmlrpcResponse struct {
	XMLName xml.Name `xml:"methodResponse"`

	Params xmlrpcParams `xml:"params"`
}

// xmlrpcParams wraps the repeated <param> children so we capture every argument. Hanging the slice off <params>
// directly matches only the one <params> element and silently drops all but the last argument.
type xmlrpcParams struct {
	Param []xmlrpcParam `xml:"param"`
}

type xmlrpcParam struct {
	Value xmlrpcValue `xml:"value"`
}

type xmlrpcValue struct {
	Array  *xmlrpcArray `xml:"array,omitempty"`
	Int    int          `xml:"i8,omitempty"`
	String string       `xml:"string,omitempty"`
}

type xmlrpcArray struct {
	Data struct {
		Value []xmlrpcArrayData `xml:"value"`
	} `xml:"data"`
}

type xmlrpcArrayData struct {
	String string `xml:"string"`
}
