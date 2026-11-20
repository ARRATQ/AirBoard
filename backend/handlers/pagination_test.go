package handlers

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func contextWithQuery(rawQuery string) *gin.Context {
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("GET", "/?"+rawQuery, nil)
	return c
}

func TestParsePageParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limits := pageLimits{Default: 20, Max: 100}

	tests := []struct {
		name       string
		query      string
		wantPage   int
		wantSize   int
		wantOffset int
	}{
		{"defaults", "", 1, 20, 0},
		{"page and page_size", "page=3&page_size=10", 3, 10, 20},
		{"limit alias", "page=2&limit=15", 2, 15, 15},
		{"page_size wins over limit", "page_size=5&limit=50", 1, 5, 0},
		{"size capped", "page_size=5000", 1, 100, 0},
		{"invalid values fall back", "page=abc&page_size=-4", 1, 20, 0},
		{"zero page", "page=0", 1, 20, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := parsePageParams(contextWithQuery(tt.query), limits)
			if p.Page != tt.wantPage || p.Size != tt.wantSize || p.Offset() != tt.wantOffset {
				t.Fatalf("got page=%d size=%d offset=%d, want %d %d %d",
					p.Page, p.Size, p.Offset(), tt.wantPage, tt.wantSize, tt.wantOffset)
			}
		})
	}
}

func TestPageParamsTotalPages(t *testing.T) {
	p := pageParams{Page: 1, Size: 10}
	for total, want := range map[int64]int{0: 0, 1: 1, 10: 1, 11: 2, 95: 10} {
		if got := p.TotalPages(total); got != want {
			t.Errorf("TotalPages(%d) = %d, want %d", total, got, want)
		}
	}
}

func TestPaginatedResponse(t *testing.T) {
	resp := pageParams{Page: 2, Size: 10}.Response([]int{1, 2}, 12)
	if resp.Page != 2 || resp.PageSize != 10 || resp.Total != 12 || resp.TotalPages != 2 {
		t.Fatalf("unexpected response %+v", resp)
	}
}
