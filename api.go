// http api

package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hashicorp/golang-lru"
	blb "github.com/matheusfillipe/blackbeard/blb"
	"github.com/matheusfillipe/blackbeard/providers"
)

const (
	CACHE_EXPIRATION = 60 * 60
	CACHE_SIZE       = 1024
)

type apiError struct {
	Error   bool   `json:"error"`
	Message string `json:"message"`
}

func apiErrorOut(c *gin.Context, err string) {
	c.JSON(http.StatusBadRequest, apiError{
		Error:   true,
		Message: err,
	})
	return
}

func getProvider(c *gin.Context, providers map[string]blb.VideoProvider) (blb.VideoProvider, bool) {
	providerName := c.Query("provider")
	if providerName == "" {
		apiErrorOut(c, "provider parameter is required")
		return nil, false
	}

	if _, ok := providers[providerName]; !ok {
		apiErrorOut(c, fmt.Sprintf("provider %q not found", providerName))
		return nil, false
	}

	return providers[providerName], true
}

type CachedValue struct {
	// Time in seconds
	time int64

	// Any value to store
	value interface{}
}

func (c CachedValue) New(value any) interface{} {
	return CachedValue{
		time:  time.Now().Unix(),
		value: value,
	}
}

func (c CachedValue) Get() (interface{}, bool) {
	if time.Now().Unix()-c.time > CACHE_EXPIRATION {
		return nil, false
	}
	return c.value, true
}

type CachedMap struct {
	// Takes a search query.lower as key
	response_shows_cache *lru.Cache
	// Takes the show url as key and show as value
	show_cache *lru.Cache
	// Takes a show url as key
	response_episodes_cache *lru.Cache
	// Takes the episode url as key and episode as value
	episode_cache *lru.Cache
	// Takes an episode url url as key
	resposne_videos_cache *lru.Cache
}

type CacheKey struct {
	a string
	b string
	c string
	d string
}

func newLruCache() *lru.Cache {
	l, _ := lru.New(CACHE_SIZE)
	return l
}

func startApiServer(host string, port int) {
	r := gin.Default()

	r.SetTrustedProxies([]string{"127.0.0.1"})

	providers := providers.GetProviders()
	providerNames := blb.Keys(providers)
	cache := make(map[string]CachedMap)

	// Each provider should have its cache
	for _, name := range providerNames {
		// Create cache map
		cache[name] = CachedMap{
			response_shows_cache:    newLruCache(),
			response_episodes_cache: newLruCache(),
			resposne_videos_cache:   newLruCache(),
			show_cache:              newLruCache(),
			episode_cache:           newLruCache(),
		}
	}

	// Life check
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Hi from blackbeard api. If you see this it means it is working!")
	})

	// Check version
	r.GET("/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"version": Version,
			"date":    BuildDate,
		})
	})

	// Providers
	type Res struct {
		Name string
		Info blb.ProviderInfo
	}
	var res []Res
	for name, prov := range providers {
		res = append(res,
			Res{
				Name: name,
				Info: prov.Info(),
			},
		)
	}
	r.GET("/providers", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"providers": res,
		})
	})

	// Show Search
	r.GET("/search", func(c *gin.Context) {
		provider, ok := getProvider(c, providers)
		if !ok {
			return
		}
		query := c.Query("q")
		if query == "" {
			apiErrorOut(c, "query parameter ('q') is required")
			return
		}
		providerName := c.Query("provider")

		// Check cache
		cache_key := strings.ToLower(query)
		if cacheMap, ok := cache[providerName]; ok {
			if cached, ok := cacheMap.response_shows_cache.Get(cache_key); ok {
				if value, ok := cached.(CachedValue).Get(); ok {
					c.JSON(http.StatusOK, value)
					return
				}
			}
		}

		shows := provider.SearchShows(query)
		response := gin.H{
			"shows": shows,
		}
		c.JSON(http.StatusOK, response)

		// Store on cache
		cache[providerName].response_shows_cache.Add(cache_key, CachedValue{}.New(response))
		for _, show := range shows {
			cache[providerName].show_cache.Add(show.Url, CachedValue{}.New(show))
		}
	})

	// Get show episodes
	r.GET("/episodes", func(c *gin.Context) {
		provider, ok := getProvider(c, providers)
		if !ok {
			return
		}
		showurl := c.Query("showurl")
		if showurl == "" {
			apiErrorOut(c, "showurl parameter is required")
			return
		}
		providerName := c.Query("provider")

		// Check cache
		cache_key := showurl
		if cacheMap, ok := cache[providerName]; ok {
			if cached, ok := cacheMap.response_shows_cache.Get(cache_key); ok {
				if value, ok := cached.(CachedValue).Get(); ok {
					c.JSON(http.StatusOK, value)
					return
				}
			}
		}

		// Check for cached show
		show := blb.Show{Url: showurl}
		if cacheMap, ok := cache[providerName]; ok {
			if cached, ok := cacheMap.show_cache.Get(showurl); ok {
				if value, ok := cached.(CachedValue).Get(); ok {
					show = value.(blb.Show)
				}
			}
		}

		episodes := provider.GetEpisodes(&show)
		response := gin.H{"episodes": episodes}
		c.JSON(http.StatusOK, response)

		// Store on cache
		cache[providerName].response_episodes_cache.Add(cache_key, CachedValue{}.New(response))
		for _, episode := range episodes {
			cache[providerName].episode_cache.Add(episode.Url, CachedValue{}.New(episode))
		}
	})

	// Get video from episode
	r.GET("/video", func(c *gin.Context) {
		provider, ok := getProvider(c, providers)
		if !ok {
			return
		}
		epurl := c.Query("epurl")
		if epurl == "" {
			apiErrorOut(c, "epurl parameter is required")
			return
		}
		providerName := c.Query("provider")

		// Check cache
		cache_key := epurl
		if cacheMap, ok := cache[providerName]; ok {
			if cached, ok := cacheMap.response_shows_cache.Get(cache_key); ok {
				if value, ok := cached.(CachedValue).Get(); ok {
					c.JSON(http.StatusOK, value)
					return
				}
			}
		}

		// Check for cached episode
		episode := blb.Episode{Url: epurl}
		if cacheMap, ok := cache[providerName]; ok {
			if cached, ok := cacheMap.episode_cache.Get(epurl); ok {
				if value, ok := cached.(CachedValue).Get(); ok {
					episode = value.(blb.Episode)
				}
			}
		}

		video := provider.GetVideo(&episode)
		c.JSON(http.StatusOK, video)

		// Store on cache
		cache[providerName].resposne_videos_cache.Add(cache_key, CachedValue{}.New(video))
	})

	r.Run(fmt.Sprintf("%v:%v", host, port))
	os.Exit(0)
}
