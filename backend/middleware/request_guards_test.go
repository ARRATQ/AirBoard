package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ulule/limiter/v3"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRateLimitKeyDoesNotPanicOnUintUserID(t *testing.T) {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	c.Set("user_id", uint(7))

	if got := rateLimitKey(c); got != "user:7" {
		t.Errorf("rateLimitKey() = %q, want %q", got, "user:7")
	}
}

func TestRateLimitBlocksAfterLimit(t *testing.T) {
	router := gin.New()
	router.POST("/login", RateLimitMiddleware(limiter.Rate{Period: time.Minute, Limit: 2}), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	codes := make([]int, 0, 3)
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/login", nil)
		req.RemoteAddr = "203.0.113.5:1234"
		router.ServeHTTP(w, req)
		codes = append(codes, w.Code)
		if i == 0 && w.Header().Get("X-RateLimit-Limit") != "2" {
			t.Errorf("X-RateLimit-Limit = %q, want \"2\"", w.Header().Get("X-RateLimit-Limit"))
		}
	}

	want := []int{http.StatusOK, http.StatusOK, http.StatusTooManyRequests}
	for i := range want {
		if codes[i] != want[i] {
			t.Fatalf("status codes = %v, want %v", codes, want)
		}
	}
}

func TestLimitRequestBody(t *testing.T) {
	router := gin.New()
	router.Use(LimitRequestBody())
	router.POST("/echo", func(c *gin.Context) {
		if _, err := io.ReadAll(c.Request.Body); err != nil {
			c.Status(http.StatusRequestEntityTooLarge)
			return
		}
		c.Status(http.StatusOK)
	})

	tests := []struct {
		name        string
		contentType string
		size        int64
		hideLength  bool
		wantStatus  int
	}{
		{"small JSON passes", "application/json", 1024, false, http.StatusOK},
		{"JSON over limit is rejected by Content-Length", "application/json", MaxJSONBodyBytes + 1, false, http.StatusRequestEntityTooLarge},
		{"JSON over limit without Content-Length is cut", "application/json", MaxJSONBodyBytes + 1, true, http.StatusRequestEntityTooLarge},
		{"upload under the upload limit passes", "multipart/form-data; boundary=x", MaxJSONBodyBytes + 1, false, http.StatusOK},
		{"upload over the upload limit is rejected", "multipart/form-data; boundary=x", MaxUploadBodyBytes + 1, false, http.StatusRequestEntityTooLarge},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := bytes.Repeat([]byte("a"), int(tt.size))
			req := httptest.NewRequest(http.MethodPost, "/echo", bytes.NewReader(body))
			req.Header.Set("Content-Type", tt.contentType)
			if tt.hideLength {
				req.ContentLength = -1
			}
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
