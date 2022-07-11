package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCustomNoQuery(t *testing.T) {
	recorder := httptest.NewRecorder()

	req := httptest.NewRequest("GET", "http://127.0.0.1/custom", nil)
	c, _ := gin.CreateTestContext(recorder)
	c.Request = req

	custom := Custom(Clients{})
	custom(c)

	if !c.IsAborted() {
		t.Error("expected request with no request parameter to be aborted")
	}
}

func TestCustomInvalidJSON(t *testing.T) {
	recorder := httptest.NewRecorder()

	req := httptest.NewRequest("GET", "http://127.0.0.1/custom?request={", nil)
	c, _ := gin.CreateTestContext(recorder)
	c.Request = req

	custom := Custom(Clients{})
	custom(c)

	if !c.IsAborted() {
		t.Error("expected request with invalid parameter to be aborted")
	}
}

func TestCustomInvalidURL(t *testing.T) {
	recorder := httptest.NewRecorder()

	req := httptest.NewRequest("GET", `http://127.0.0.1/custom?request={"method":">"}`, nil)
	c, _ := gin.CreateTestContext(recorder)
	c.Request = req

	custom := Custom(Clients{})
	custom(c)

	if !c.IsAborted() {
		t.Error("expected request with invalid HTTP method to be aborted")
	}
}

func TestCustomInvalidHost(t *testing.T) {
	recorder := httptest.NewRecorder()
	url := `http://127.0.0.1/custom?request={"url":"http://_._._._._.com"}`

	req := httptest.NewRequest("GET", url, nil)
	c, _ := gin.CreateTestContext(recorder)
	c.Request = req

	custom := Custom(Clients{&http.Client{}, &http.Client{}})
	custom(c)

	if !c.IsAborted() {
		t.Error("expected custom request to be aborted")
	}
}

func TestCustomValidRequest(t *testing.T) {
	recorder := httptest.NewRecorder()
	url := `http://127.0.0.1/custom?request={"url":"http://www.example.com","headers":{"x-test":"1"},"cookies":{"x-test":"1"}}`

	for _, ext := range []string{"", "&noproxy"} {
		req := httptest.NewRequest("GET", url+ext, nil)
		c, _ := gin.CreateTestContext(recorder)
		c.Request = req

		custom := Custom(Clients{&http.Client{}, &http.Client{}})
		custom(c)

		if c.IsAborted() {
			t.Error("expected custom request to be performed successfully")
		}
	}
}

func TestCustomValidRequestBody(t *testing.T) {
	recorder := httptest.NewRecorder()
	url := `http://127.0.0.1/custom?request={"url":"http://www.example.com","body":"test"}`

	req := httptest.NewRequest("GET", url, nil)
	c, _ := gin.CreateTestContext(recorder)
	c.Request = req

	custom := Custom(Clients{&http.Client{}, &http.Client{}})
	custom(c)

	if c.IsAborted() {
		t.Error("expected custom request to be performed successfully")
	}
}
