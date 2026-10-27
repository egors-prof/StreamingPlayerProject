package httpx

import (
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const requestIDHeader = "X-Request-ID"

// ErrorResponse — единый формат ошибки для всех сервисов.
type ErrorResponse struct {
	Error string `json:"error"`
}

// Fail прерывает обработку запроса и отдаёт ошибку в едином формате.
func Fail(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, ErrorResponse{Error: msg})
}

// RequestID берёт X-Request-ID от шлюза или генерирует новый. По нему можно
// найти в логах все записи одного запроса.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.GetHeader(requestIDHeader)
		if id == "" || len(id) > 64 {
			b := make([]byte, 8)
			_, _ = rand.Read(b)
			id = hex.EncodeToString(b)
		}
		c.Set("request_id", id)
		c.Header(requestIDHeader, id)
		c.Next()
	}
}

// AccessLog пишет одну строку на запрос. Логируется только путь без query:
// в query лежит билет на стрим, а токены в логи попадать не должны.
func AccessLog(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		log.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"ip", c.ClientIP(),
			"request_id", c.GetString("request_id"),
		)
	}
}

// Recovery превращает панику в ответ 500, не раскрывая клиенту детали.
func Recovery(log *slog.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		log.Error("panic recovered", "error", err, "request_id", c.GetString("request_id"))
		Fail(c, http.StatusInternalServerError, "internal error")
	})
}

// RateLimit ограничивает число запросов с одного IP за окно времени
// (fixed window). Нужен для защиты входа от перебора паролей.
// Хранится в памяти процесса: при нескольких репликах нужен Redis.
func RateLimit(limit int, window time.Duration) gin.HandlerFunc {
	type counter struct {
		count int
		reset time.Time
	}
	var (
		mu       sync.Mutex
		counters = make(map[string]*counter)
		lastGC   = time.Now()
	)
	return func(c *gin.Context) {
		now := time.Now()
		ip := c.ClientIP()

		mu.Lock()
		if now.Sub(lastGC) > window { // чистим старые записи, чтобы map не росла бесконечно
			for k, v := range counters {
				if now.After(v.reset) {
					delete(counters, k)
				}
			}
			lastGC = now
		}
		ctr, ok := counters[ip]
		if !ok || now.After(ctr.reset) {
			ctr = &counter{reset: now.Add(window)}
			counters[ip] = ctr
		}
		ctr.count++
		over := ctr.count > limit
		retryAfter := int(ctr.reset.Sub(now).Seconds()) + 1
		mu.Unlock()

		if over {
			c.Header("Retry-After", strconv.Itoa(retryAfter))
			Fail(c, http.StatusTooManyRequests, "too many requests, try again later")
			return
		}
		c.Next()
	}
}
