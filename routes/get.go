package routes

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// Get simplifies performing a custom GET request
func Get(clients Clients) gin.HandlerFunc {
	custom := Custom(clients)
	return func(c *gin.Context) {

		// Make sure url parameter was specified
		param := c.Query("url")
		if param == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":  "url not specified",
				"status": http.StatusBadRequest,
			})
			return
		}

		// Create custom request
		creq := CustomRequest{
			Method: "GET",
			URL:    param,
		}

		// Marshal request into JSON string
		req, err := json.Marshal(creq)
		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":  "error encoding json",
				"status": http.StatusInternalServerError,
			})
			return
		}

		// Set query string
		query := url.Values{
			"request": []string{string(req)},
		}
		// Check for noproxy
		if _, noproxy := c.GetQuery("noproxy"); noproxy {
			query.Set("noproxy", "")
		}
		// Check for stream
		if _, stream := c.GetQuery("stream"); stream {
			query.Set("stream", "")
		}
		c.Request.URL.RawQuery = query.Encode()

		// Call custom route
		custom(c)
	}
}
