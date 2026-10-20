package cookie

import (
	"net/http"
	"strings"
	"time"

	"backend/internal/config"

	"github.com/gin-gonic/gin"
)

func SetRefreshToken(c *gin.Context, cfg config.CookieConfig, value string, expiresAt time.Time) {
	maxAge := int(time.Until(expiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	//	set the same site and cookie
	c.SetSameSite(parseSameSite(cfg.SameSite))
	c.SetCookie(cfg.RefreshName, value, maxAge, cfg.Path, "", cfg.Secure, true)
}

func ClearRefreshToken(c *gin.Context, cfg config.CookieConfig) {
	//	clear the same site and cookie
	c.SetSameSite(parseSameSite(cfg.SameSite))
	c.SetCookie(cfg.RefreshName, "", -1, cfg.Path, "", cfg.Secure, true)
}

func RefreshTokenFromRequest(c *gin.Context, cfg config.CookieConfig) (string, error) {
	//	get the refresh token from the request
	token, err := c.Cookie(cfg.RefreshName)
	if err != nil {
		if err == http.ErrNoCookie {
			return "", err
		}
		return "", err
	}

	//	check if the token is empty
	if token == "" {
		return "", http.ErrNoCookie
	}
	return token, nil
}

func parseSameSite(v string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
