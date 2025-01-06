package middleware

import (
	"bytes"
	"context"
	"io"
	"time"

	"github.com/gin-gonic/gin"
)

type LogMiddlewareBuilder struct {
	logFunc       func(ctx context.Context, accessLog AccessLog)
	allowReqBody  bool
	allowRespBody bool
	maxPathLength int
	maxBodyLength int
}

func NewLogMiddlewareBuilder(logFunc func(ctx context.Context, accessLog AccessLog)) *LogMiddlewareBuilder {
	return &LogMiddlewareBuilder{
		logFunc: logFunc,
	}
}

func (l *LogMiddlewareBuilder) AllowReqBody() *LogMiddlewareBuilder {
	l.allowReqBody = true
	return l
}

func (l *LogMiddlewareBuilder) AllowRespBody() *LogMiddlewareBuilder {
	l.allowRespBody = true
	return l
}

func (l *LogMiddlewareBuilder) SetMaxPathLength(maxPathLength int) *LogMiddlewareBuilder {
	l.maxPathLength = maxPathLength
	return l
}

func (l *LogMiddlewareBuilder) SetMaxBodyLength(maxBodyLength int) *LogMiddlewareBuilder {
	l.maxBodyLength = maxBodyLength
	return l
}

func (l *LogMiddlewareBuilder) DefaultMaxPathLength() *LogMiddlewareBuilder {
	return l.SetMaxPathLength(1024)
}

func (l *LogMiddlewareBuilder) DefaultMaxBodyLength() *LogMiddlewareBuilder {
	return l.SetMaxBodyLength(4096)
}

func (l *LogMiddlewareBuilder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		path := ctx.Request.URL.Path
		if len(path) > l.maxPathLength {
			path = path[:l.maxPathLength] + "..."
		}
		method := ctx.Request.Method
		accessLog := AccessLog{
			Path:   path,
			Method: method,
		}
		if l.allowReqBody {
			body, _ := ctx.GetRawData()
			if len(body) > l.maxBodyLength {
				accessLog.ReqBody = string(body[:l.maxBodyLength]) + "..."
			} else {
				accessLog.ReqBody = string(body)
			}
			ctx.Request.Body = io.NopCloser(bytes.NewReader(body))
		}

		start := time.Now()

		if l.allowRespBody {
			ctx.Writer = &responseWriter{
				ResponseWriter: ctx.Writer,
				accessLog:      &accessLog,
			}
		}

		defer func() {
			accessLog.Duration = time.Since(start)
			l.logFunc(ctx, accessLog)
		}()

		ctx.Next()
	}
}

type AccessLog struct {
	Path     string        `json:"path"`
	Method   string        `json:"method"`
	ReqBody  string        `json:"req_body"`
	RespBody string        `json:"resp_body"`
	Status   int           `json:"status"`
	Duration time.Duration `json:"duration"`
}

type responseWriter struct {
	gin.ResponseWriter
	accessLog *AccessLog
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.accessLog.RespBody = string(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.accessLog.Status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
