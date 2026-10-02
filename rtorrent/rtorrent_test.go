package rtorrent

import (
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/kolo/xmlrpc"
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

func TestMulticallByHash(t *testing.T) {
	t.Parallel()

	const (
		okReply = `<?xml version="1.0"?><methodResponse><params><param><value><array><data>
<value><array><data><value><string>a name</string></value></data></array></value>
<value><array><data><value><i8>1</i8></value></data></array></value>
</data></array></value></param></params></methodResponse>`
		faultReply = `<?xml version="1.0"?><methodResponse><params><param><value><array><data>
<value><array><data><value><string>a name</string></value></data></array></value>
<value><struct><member><name>faultCode</name><value><i4>-501</i4></value></member>
<member><name>faultString</name><value><string>Could not find info-hash.</string></value></member></struct></value>
</data></array></value></param></params></methodResponse>`
		shortReply = `<?xml version="1.0"?><methodResponse><params><param><value><array><data>
<value><array><data><value><string>a name</string></value></data></array></value>
</data></array></value></param></params></methodResponse>`
		structReply = `<?xml version="1.0"?><methodResponse><params><param><value><array><data>
<value><array><data><value><string>a name</string></value></data></array></value>
<value><struct><member><name>foo</name><value><string>bar</string></value></member></struct></value>
</data></array></value></param></params></methodResponse>`
	)

	tests := []struct {
		name      string
		reply     string
		want      []any
		wantErr   error
		wantFault string
	}{
		{name: "values are unwrapped in order", reply: okReply, want: []any{testName, int64(1)}},
		{name: "a fault on one method fails the call", reply: faultReply, wantFault: "Could not find info-hash."},
		{name: "result count mismatch", reply: shortReply, wantErr: ErrBadData},
		{name: "a struct without faultCode is not a fault", reply: structReply, wantErr: ErrBadData},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("reading request: %v", err)
					return
				}
				for _, want := range []string{
					"<methodName>system.multicall</methodName>",
					"<string>d.name</string>",
					"<string>d.complete</string>",
					"<string>" + testInfoHash + "</string>",
				} {
					assert.Contains(t, string(body), want)
				}
				_, _ = io.WriteString(w, tt.reply)
			}))
			t.Cleanup(s.Close)

			c, err := New(s.URL, nil)
			require.NoError(t, err)
			t.Cleanup(func() { _ = c.Close() })

			got, err := c.multicallByHash(testInfoHash, testMethod, "d.complete=")
			switch {
			case tt.wantFault != "":
				var fault xmlrpc.FaultError
				require.ErrorAs(t, err, &fault)
				assert.Equal(t, tt.wantFault, fault.String)
				assert.Equal(t, -501, fault.Code)
			case tt.wantErr != nil:
				require.ErrorIs(t, err, tt.wantErr)
			default:
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestMulticallEntryFor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command string
		want    multicallEntry
	}{
		{"bare", testMethod, multicallEntry{MethodName: testMethod, Params: []any{testInfoHash}}},
		{"trailing equals", "d.name=", multicallEntry{MethodName: testMethod, Params: []any{testInfoHash}}},
		{"one arg", "d.custom=label", multicallEntry{MethodName: "d.custom", Params: []any{testInfoHash, "label"}}},
		{
			"comma separated args", "d.custom.set=key,value",
			multicallEntry{MethodName: "d.custom.set", Params: []any{testInfoHash, "key", "value"}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, multicallEntryFor(testInfoHash, tt.command))
		})
	}
}

func TestMulticallByHashKeepsMethodNameFirst(t *testing.T) {
	t.Parallel()

	methods := slices.Repeat([]string{testMethod}, 64)
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request: %v", err)
			return
		}
		assert.Equal(t, len(methods), strings.Count(string(body), "<struct><member><name>methodName</name>"),
			"every entry must lead with methodName")
		http.Error(w, "done", http.StatusInternalServerError)
	}))
	t.Cleanup(s.Close)

	c, err := New(s.URL, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = c.Close() })

	_, err = c.multicallByHash(testInfoHash, methods...)
	require.Error(t, err)
}

func TestMulticallByHashRequiresInfoHash(t *testing.T) {
	t.Parallel()

	_, err := (&XMLRPCClient{}).multicallByHash("", testMethod)
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
