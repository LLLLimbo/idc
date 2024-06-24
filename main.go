package main

import (
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"slices"
	"time"
)

var (
	apps         = flag.String("apps", "apps.json", "apps config file")
	host         = flag.String("host", "http://localhost:29705", "proxy host")
	keycloak     = flag.String("keycloak", "http://keycloak:9998/realms/master", "keycloak URL")
	clientId     = flag.String("client-id", "idc", "client ID")
	clientSecret = flag.String("client-secret", "FxotKy0hWrkVzzlIPXW4q5edc5GK8788", "client secret")
	authCenter   = flag.String("auth-center", "http://localhost:29706", "auth center URL")
)

func main() {
	flag.Parse()
	r := gin.Default()
	r.Use(CORSMiddleware())
	cache := NewCache()
	ac := NewAppContext()
	ac.ReadApps()

	r.GET("/idc/redirect/login", RedirectLoginHandler(ac, cache))

	r.GET("/idc/redirect/callback", RedirectCallbackHandler(cache, ac))

	r.GET("/idc/session/fetch/:key/:appId", SessionFetch(cache))

	err := r.Run("0.0.0.0:29705")
	if err != nil {
		panic(err)
	}
}

func SessionFetch(cache *Cache) func(c *gin.Context) {
	return func(c *gin.Context) {
		otk := c.Param("key")

		key := fmt.Sprintf("t_%s", otk)
		token, exists := cache.Get(key)
		if exists {
			cache.DeleteItem(key)
			c.JSON(http.StatusOK, gin.H{"token": token})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "your code has expired"})
	}
}

func RedirectCallbackHandler(cache *Cache, ac *AppContext) func(c *gin.Context) {
	return func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")
		sessionState := c.Query("session_state")
		_, token, err := MakeTokenRequest(state, code)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error request token"})
			c.Abort()
			return
		}

		//get the app config from the cache
		app := &AppConfig{}
		appId, cacheExists := cache.Get(state)
		if cacheExists {
			app = ac.GetAppConfig(appId)
		}

		_, cred := CreateSession(token, sessionState)
		log.Printf("Setting token cache for app %s", appId)
		otk := randomStr(6)
		cache.Set(fmt.Sprintf("t_%s", otk), cred.ToJsonStr(), 30*time.Minute)
		//redirect to the app's token endpoint
		isCookieType := slices.Contains(app.SessionType, "cookie")
		if isCookieType {
			c.SetCookie("session_id", cred.Id, cred.ExpiresIn, "/", "localhost", false, true)
			c.Redirect(http.StatusFound, app.Redirect)
			c.Abort()
			return
		}
		c.SetCookie("session_id", cred.Id, cred.ExpiresIn, "/", "localhost", false, true)
		c.Redirect(http.StatusFound, fmt.Sprintf("%s?key=%s", app.TokenEndpoint, otk))
	}
}

func RedirectLoginHandler(ac *AppContext, cache *Cache) func(c *gin.Context) {
	return func(c *gin.Context) {
		state := c.Query("state")
		if state == "" {
			state = randomStr(6)
		}
		appId := c.Query("app_id")

		//validate appid
		if appId == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "app id is required"})
			c.Abort()
			return
		}
		app := &AppConfig{}
		app = ac.GetAppConfig(appId)
		if app == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "app id is required"})
			c.Abort()
			return
		}

		callback := fmt.Sprintf("%s/idc/redirect/callback", *host)
		uri := fmt.Sprintf("%s/protocol/openid-connect/auth?client_id=%s&response_type=code&redirect_uri=%s&scope=openid profile&state=%s", *keycloak, *clientId, callback, state)

		cache.Set(state, appId, 30*time.Minute)
		c.JSON(http.StatusOK, gin.H{"redirect": uri})
	}
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
