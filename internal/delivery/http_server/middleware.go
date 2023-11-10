package apiserver

import (
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func (s *APIServer) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authorizationHeader := ctx.GetHeader("Authorization")
		if authorizationHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			ctx.Abort()
			return
		}

		tokenString := strings.Split(authorizationHeader, " ")[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte(s.config.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			ctx.Abort()
			return
		}

		ctx.Set("userID", claims["sub"])
		ctx.Next()
	}
}

func (s *APIServer) CacheMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.Request.URL.String()

		if value, err := s.cache.Get(key); err == nil {
			c.String(http.StatusOK, value)
			return
		}

		c.Next()

		if c.Writer.Status() == http.StatusOK {
			if body, err := c.GetRawData(); err != nil {
				if err := s.cache.Set(key, string(body), 0); err != nil {
					s.logger.Warn(err)
				}
			}
		}
	}
}
