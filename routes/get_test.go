package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestGet(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	get := Get(Clients{&http.Client{}, &http.Client{}})

	url := "http://127.0.0.1/get?url=http://www.example.com"

	for _, ext := range []string{"", "&noproxy"} {
		c.Request = httptest.NewRequest("GET", url+ext, nil)
		get(c)

		if c.IsAborted() {
			t.Errorf("expected request not to be aborted")
		}
	}
}

func TestGetNoQuery(t *testing.T) {
	recorder := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(recorder)

	get := Get(Clients{&http.Client{}, &http.Client{}})

	c.Request = httptest.NewRequest("GET", "http://127.0.0.1/get?url=", nil)
	get(c)

	if !c.IsAborted() {
		t.Errorf("expected request to be aborted")
	}
}
