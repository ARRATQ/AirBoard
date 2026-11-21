package handlers

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

func contextWithBody(body string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	return c
}

func TestExplicitFalseColumns(t *testing.T) {
	tests := []struct {
		name    string
		body    string
		columns []string
		want    map[string]interface{}
	}{
		{"false envoyé", `{"name":"x","is_active":false}`, []string{"is_active"}, map[string]interface{}{"is_active": false}},
		{"true envoyé", `{"is_active":true}`, []string{"is_active"}, map[string]interface{}{}},
		{"champ absent", `{"name":"x"}`, []string{"is_active"}, map[string]interface{}{}},
		{"null", `{"is_active":null}`, []string{"is_active"}, map[string]interface{}{}},
		{"plusieurs colonnes", `{"is_active":false,"open_in_new_tab":false,"other":false}`,
			[]string{"is_active", "open_in_new_tab"}, map[string]interface{}{"is_active": false, "open_in_new_tab": false}},
		{"corps invalide", `pas du json`, []string{"is_active"}, map[string]interface{}{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := explicitFalseColumns(contextWithBody(tt.body), tt.columns...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("explicitFalseColumns = %v, attendu %v", got, tt.want)
			}
		})
	}
}

func TestExplicitFalseColumnsAfterBinding(t *testing.T) {
	c := contextWithBody(`{"name":"x","is_active":false}`)
	var req struct {
		Name     string `json:"name"`
		IsActive bool   `json:"is_active"`
	}
	if err := c.ShouldBindBodyWith(&req, binding.JSON); err != nil {
		t.Fatalf("liaison : %v", err)
	}
	if got := explicitFalseColumns(c, "is_active"); !reflect.DeepEqual(got, map[string]interface{}{"is_active": false}) {
		t.Errorf("après liaison = %v", got)
	}
}
