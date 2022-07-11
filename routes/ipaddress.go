package routes

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// IPAddress returns a gin.HandlerFunc to get the IP address of the server
func IPAddress(clients Clients) gin.HandlerFunc {
	return func(c *gin.Context) {

		// Get appropriate client
		var client *http.Client
		if _, noproxy := c.GetQuery("noproxy"); noproxy {
			client = clients.Normal
		} else {
			client = clients.Proxy
		}

		// Perform request
		resp, err := client.Get("https://icanhazip.com")
		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":  fmt.Sprintf("%v", err),
				"status": http.StatusInternalServerError,
			})
			return
		}

		defer resp.Body.Close()

		// Read response's body
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":  fmt.Sprintf("%v", err),
				"status": http.StatusInternalServerError,
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": http.StatusOK,
			"data":   strings.TrimSpace(string(body)),
		})
	}
}
