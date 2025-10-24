package apikey

import (
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
	"github.com/virzz/mulan/rsp"
	"github.com/virzz/mulan/rsp/code"
)

type Config struct {
	Name    string   `json:"name" yaml:"name"`
	Secret  string   `json:"secret" yaml:"secret"`
	Secrets []string `json:"secrets" yaml:"secrets"`
}

func FlagSet(defaultPort int) *pflag.FlagSet {
	fs := pflag.NewFlagSet("http", pflag.ContinueOnError)
	fs.String("apikey.name", "apikey", "APIKey Name")
	fs.String("apikey.secret", "", "APIKey Secret")
	fs.StringSlice("apikey.secrets", []string{}, "APIKey Secret List")
	return fs
}

func Mw(cfg *Config, skipPaths ...string) func(*gin.Context) {
	name := strings.TrimSpace(cfg.Name)
	apikeys := append([]string{cfg.Secret}, cfg.Secrets...)
	if len(apikeys) == 0 {
		return func(c *gin.Context) { c.Next() }
	}
	skipPathMap := make(map[string]struct{})
	for _, path := range skipPaths {
		skipPathMap[path] = struct{}{}
	}
	return func(c *gin.Context) {
		if _, ok := skipPathMap[c.Request.URL.Path]; ok {
			c.Next()
			return
		}
		current := c.GetHeader("apikey")
		if current == "" {
			current = strings.TrimSpace(c.GetHeader(name))
		}
		if current == "" {
			current, _ = c.Cookie(name)
			current = strings.TrimSpace(current)
		}
		if current == "" {
			current = strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		}
		if current != "" && slices.Contains(apikeys, current) {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(401, rsp.E(code.Unauthorized, "Error Unauthorized"))
	}
}
