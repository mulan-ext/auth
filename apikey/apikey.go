package apikey

import (
	"crypto/subtle"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/pflag"
)

type Config struct {
	Name   string   `json:"name" yaml:"name"`
	Value  string   `json:"value" yaml:"value"`
	Values []string `json:"values" yaml:"values"`
}

func (c *Config) FlagSet() *pflag.FlagSet { return FlagSet() }

func FlagSet() *pflag.FlagSet {
	fs := pflag.NewFlagSet("apikey", pflag.ContinueOnError)
	fs.String("apikey.name", "apikey", "APIKey Name")
	fs.String("apikey.value", "", "APIKey Value")
	fs.StringSlice("apikey.values", []string{}, "APIKey Value List")
	return fs
}

func Mw(cfg *Config, skipPaths ...string) func(*gin.Context) {
	name := strings.TrimSpace(cfg.Name)
	apikeys := append([]string{cfg.Value}, cfg.Values...)
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
		if current == "" && name != "" {
			current = strings.TrimSpace(c.GetHeader(name))
			if current == "" {
				current, _ = c.Cookie(name)
				current = strings.TrimSpace(current)
			}
		}
		if current == "" {
			current = strings.TrimSpace(strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer "))
		}
		if current != "" {
			for _, key := range apikeys {
				if key == "" {
					continue
				}
				// 使用常量时间比较防止计时攻击
				if len(key) == len(current) && subtle.ConstantTimeCompare([]byte(key), []byte(current)) == 1 {
					c.Next()
					return
				}
			}
		}
		c.AbortWithStatus(401)
	}
}
