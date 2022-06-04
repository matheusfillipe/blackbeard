// TODO http api

package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	blb "github.com/matheusfillipe/blackbeard/blb"
	"github.com/matheusfillipe/blackbeard/providers"
)

type apiError struct {
	Error   bool `json:"error"`
	Message string `json:"message"`
}

func apiErrorOut(c *gin.Context, err string) {
      c.JSON(http.StatusBadRequest, apiError{
        Error:   true,
        Message: err,
      })
      return
}

func startApiServer(host string, port int) {
	r := gin.Default()

	r.SetTrustedProxies([]string{"127.0.0.1"})

	providers := providers.GetProviders()
	providerNames := blb.Keys(providers)

	// Life check
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hi from blackbeard api")
	})

	// Check version
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version": Version,
			"date":    BuildDate,
		})
	})

	// Providers
	r.GET("/providers", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"providers": providerNames,
		})
	})

	// Show Search
	r.GET("/search", func(c *gin.Context) {
    providerName := c.Query("provider")
    if providerName == "" {
      apiErrorOut(c, "provider parameter is required")
      return
    }

    if _, ok := providers[providerName]; !ok {
      apiErrorOut(c, fmt.Sprintf("provider %q not found", providerName))
      return
    }

		provider := providers[providerName]
    query := c.Query("q")
    if query == "" {
      apiErrorOut(c, "query parameter ('q') is required")
      return
    }
    shows := provider.SearchShows(query)
		c.JSON(http.StatusOK, shows)
	})

  // Get show episodes

	r.Run(fmt.Sprintf("%v:%v", host, port))
	os.Exit(0)
}
