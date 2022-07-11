package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestIPAddress(t *testing.T) {
	recorder := httptest.NewRecorder()

	ip := IPAddress(Clients{&http.Client{}, &http.Client{}})
	c, _ := gin.CreateTestContext(recorder)
	url := "http://127.0.0.1/get"

	for _, ext := range []string{"", "noproxy"} {
		c.Request = httptest.NewRequest("GET", url+ext, nil)

		ip(c)

		if c.IsAborted() {
			t.Error("expected request not to be aborted")
		}
	}
}
