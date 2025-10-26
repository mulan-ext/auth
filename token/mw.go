package token

import (
	"github.com/gin-gonic/gin"
)

func RoleMW(roles ...string) gin.HandlerFunc {
	roleMap := make(map[string]struct{})
	for _, role := range roles {
		roleMap[role] = struct{}{}
	}
	return func(c *gin.Context) {
		for _, r := range c.GetStringSlice(CtxKeyRoles) {
			if _, ok := roleMap[r]; ok {
				c.Next()
				return
			}
		}
		c.AbortWithStatus(403)
	}
}

func AuthMW() gin.HandlerFunc {
	return func(c *gin.Context) {
		if _, ok := c.Get(CtxKeyID); ok {
			c.Next()
			return
		}
		c.AbortWithStatus(401)
	}
}
