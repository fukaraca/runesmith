package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"strings"

	gb "github.com/fukaraca/skypiea/pkg/guest_book"
	"github.com/gin-gonic/gin"
	sloggin "github.com/samber/slog-gin"
	"golang.org/x/time/rate"
)

type ctxKey string

const (
	LoggerCtx ctxKey = "logger"
)

func LoggerMw(logger *slog.Logger) gin.HandlerFunc {
	return sloggin.NewWithConfig(logger, sloggin.Config{
		WithUserAgent: true,
		WithRequestID: true,
		Filters:       []sloggin.Filter{doNotLogIfNoErr("/healthz")},
	})
}

func GetLoggerFromContext(ctx context.Context) *slog.Logger {
	if v, ok := ctx.Value(string(LoggerCtx)).(*slog.Logger); ok {
		return v
	}
	return slog.Default()
}

func doNotLogIfNoErr(url string) sloggin.Filter {
	return func(ctx *gin.Context) bool {
		if ctx.FullPath() == url && len(ctx.Errors) == 0 {
			return false
		}
		return true
	}
}

func ErrorHandlerMw() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if len(c.Errors) > 0 {
			GetLoggerFromContext(c).Error("request has failed", "errors", c.Errors)
		}
	}
}

func CounterUIMw() gin.HandlerFunc {
	return func(c *gin.Context) {
		rip := realIP(c.Request)
		gb.GuestBook.RegisterGuest(rip, c.Request.URL.Path)
		c.Next()
	}
}

func realIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		parts := strings.Split(xf, ",")
		return strings.TrimSpace(parts[0])
	}
	if xr := r.Header.Get("X-Real-IP"); xr != "" {
		return xr
	}
	host, _, _ := net.SplitHostPort(r.RemoteAddr)
	return host
}

func RateLimiterMw() gin.HandlerFunc {
	limiter := rate.NewLimiter(1, 10)
	return func(c *gin.Context) {

		if limiter.Allow() {
			c.Next()
		} else {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"message": "Limits exceed",
			})
		}
	}
}
