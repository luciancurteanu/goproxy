package routes

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Custom returns a gin.HandlerFunc to perform a custom HTTP request
func Custom(clients Clients) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Make sure request parameter was specified
		param := c.Query("request")
		if param == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":  "request not specified",
				"status": http.StatusBadRequest,
			})
			return
		}

		// Parse custom request
		var creq CustomRequest
		if err := json.Unmarshal([]byte(param), &creq); err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":  "error parsing json",
				"status": http.StatusBadRequest,
			})
			return
		}

		var body io.Reader
		if creq.Body != "" {
			body = strings.NewReader(creq.Body)
		}

		// Check if we should proxy the request (default is true)
		var client *http.Client
		if _, noproxy := c.GetQuery("noproxy"); noproxy {
			client = clients.Normal
		} else {
			client = clients.Proxy
		}

		// Create HTTP request to perform
		req, err := http.NewRequest(creq.Method, creq.URL, body)
		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":  fmt.Sprintf("%v", err),
				"status": http.StatusBadRequest,
			})
			return
		}

		// Set headers
		for h, val := range creq.Headers {
			req.Header.Set(h, val)
		}

		// Set cookies
		for c, val := range creq.Cookies {
			req.AddCookie(&http.Cookie{Name: c, Value: val})
		}

		// Perform the HTTP request
		log.Infof("peforming %s request to %s", req.Method, req.URL)
		resp, err := client.Do(req)
		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":  fmt.Sprintf("%v", err),
				"status": http.StatusInternalServerError,
			})
			return
		}
		log.Infof("%s %s %s", req.Method, req.URL, resp.Status)

		// Make sure we clean up...
		defer resp.Body.Close()

		html, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":  fmt.Sprintf("%v", err),
				"status": http.StatusInternalServerError,
			})
			return
		}

		// If stream is specified just send the raw bytes with the same
		// Content-Type as the response
		if _, stream := c.GetQuery("stream"); stream {
			c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), html)
			return
		}

		// Return a JSON response and forward the response code
		c.JSON(resp.StatusCode, gin.H{
			"request": gin.H{
				"url":     req.URL.String(),
				"headers": req.Header,
				"cookies": req.Cookies(),
			},
			"response": gin.H{
				"headers": resp.Header,
				"cookies": resp.Cookies(),
			},
			"status": resp.StatusCode,
			"data":   string(html),
		})
	}
}
