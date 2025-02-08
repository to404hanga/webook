package middleware

import (
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type LoginMiddlewareBuilder struct {
}

func (m *LoginMiddlewareBuilder) CheckLogin() gin.HandlerFunc {
	gob.Register(time.Now())
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if path == "/users/signup" || path == "/users/login" {
			return
		}
		sess := sessions.Default(ctx)
		userId := sess.Get("userId")
		if userId == nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		current := time.Now()

		const updateTimeKey = "updateTime"
		value := sess.Get(updateTimeKey)
		lassUpdateTime, ok := value.(time.Time)
		if !ok || current.Sub(lassUpdateTime) > time.Minute {
			sess.Set(updateTimeKey, current)
			sess.Set("userId", userId)
			if err := sess.Save(); err != nil {
				fmt.Println(err)
			}
		}
	}
}
