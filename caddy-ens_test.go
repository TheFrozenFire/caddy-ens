package caddyens

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/caddyserver/caddy/v2"
	"github.com/caddyserver/caddy/v2/modules/caddyhttp"
)

func TestHandler(t *testing.T) {
    rpc_endpoint := os.Getenv("ETH_RPC_ENDPOINT")

	for i, tc := range []struct {
		handler Handler
		reqURI  string
		expect  map[string]any
	}{
	    
	} {
		if err := tc.handler.Provision(caddy.Context{}); err != nil {
			t.Fatalf("Test %d: Provisioning handler: %v", i, err)
		}

		req, err := http.NewRequest(http.MethodGet, tc.reqURI, nil)
		if err != nil {
			t.Fatalf("Test %d: Creating request: %v", i, err)
		}
		repl := caddyhttp.NewTestReplacer(req)
		repl.Set("testvar", "testing")
		ctx := context.WithValue(req.Context(), caddy.ReplacerCtxKey, repl)
		req = req.WithContext(ctx)

		rr := httptest.NewRecorder()
		noop := caddyhttp.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) error { return nil })

		if err := tc.handler.ServeHTTP(rr, req, noop); err != nil {
			t.Errorf("Test %d: Handler returned error: %v", i, err)
			continue
		}

		for key, expected := range tc.expect {
			actual, _ := repl.Get(key)
			if !reflect.DeepEqual(actual, expected) {
				t.Errorf("Test %d: Expected %#v but got %#v for {%s}", i, expected, actual, key)
			}
		}
	}
}
