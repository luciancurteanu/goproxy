package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWhitelist(t *testing.T) {
	addresses := []string{
		"10.0.0.1",
		"127.0.0.1",
	}

	whitelist := Whitelist(addresses)
	recorder := httptest.NewRecorder()

	for _, address := range addresses {
		c, _ := gin.CreateTestContext(recorder)
		c.Request = &http.Request{
			RemoteAddr: address + ":5589",
		}
		whitelist(c)
		if c.IsAborted() {
			t.Errorf("expected request from %q not to be aborted", address)
		}
	}
	address := "10.0.0.2"
	c, _ := gin.CreateTestContext(recorder)
	c.Request = &http.Request{
		RemoteAddr: address,
	}
	whitelist(c)
	if !c.IsAborted() {
		t.Errorf("expected request from %q to be aborted", address)
	}
}

func TestHeaders(t *testing.T) {
	values := map[string]string{
		"access-control-allow-origin": "*",
		"x-device-id":                 "10923",
		"some-other-header":           "something",
	}

	headers := Headers(values)

	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)
	headers(c)

	for key, val := range values {
		if got := c.Writer.Header().Get(key); got != val {
			t.Errorf("expected header %q to be set to %q; got %q", key, val, got)
		}
	}
}
