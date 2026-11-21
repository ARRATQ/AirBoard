package apierror

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func record(t *testing.T, handler gin.HandlerFunc, body string) (*httptest.ResponseRecorder, map[string]interface{}) {
	t.Helper()
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(body))
	c.Request.Header.Set("Content-Type", "application/json")
	handler(c)
	var got map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("corps non JSON %q : %v", w.Body.String(), err)
	}
	return w, got
}

func TestRespondEnvelope(t *testing.T) {
	cases := []struct {
		status int
		code   string
	}{
		{http.StatusBadRequest, "bad_request"},
		{http.StatusUnauthorized, "unauthorized"},
		{http.StatusForbidden, "forbidden"},
		{http.StatusNotFound, "not_found"},
		{http.StatusConflict, "conflict"},
		{http.StatusRequestEntityTooLarge, "payload_too_large"},
		{http.StatusTooManyRequests, "too_many_requests"},
		{http.StatusInternalServerError, "internal_error"},
		{http.StatusBadGateway, "internal_error"},
		{http.StatusServiceUnavailable, "service_unavailable"},
	}
	for _, tc := range cases {
		w, body := record(t, func(c *gin.Context) { Respond(c, tc.status, "Message lisible") }, "")
		if w.Code != tc.status {
			t.Errorf("%d : statut %d", tc.status, w.Code)
		}
		want := map[string]interface{}{"error": "Message lisible", "code": tc.code}
		if len(body) != len(want) || body["error"] != want["error"] || body["code"] != want["code"] {
			t.Errorf("%d : corps %v, attendu %v", tc.status, body, want)
		}
	}
}

func TestRespondCodeKeepsDomainCode(t *testing.T) {
	w, body := record(t, func(c *gin.Context) {
		RespondCode(c, http.StatusForbidden, CodeCommentsDisabled, "Les commentaires sont désactivés")
	}, "")
	if w.Code != http.StatusForbidden || body["code"] != CodeCommentsDisabled || body["error"] != "Les commentaires sont désactivés" {
		t.Fatalf("statut %d, corps %v", w.Code, body)
	}
}

func TestAbortStopsChain(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	Abort(c, http.StatusUnauthorized, "Token manquant")
	if !c.IsAborted() || w.Code != http.StatusUnauthorized {
		t.Fatalf("aborted=%v statut=%d", c.IsAborted(), w.Code)
	}
}

type bindTarget struct {
	Title      string `json:"title" binding:"required"`
	Email      string `json:"email" binding:"omitempty,email"`
	CategoryID uint   `json:"category_id"`
}

func TestInvalidBodyMessages(t *testing.T) {
	cases := []struct {
		name string
		body string
		want string
	}{
		{"corps vide", "", "Données invalides : JSON absent ou mal formé"},
		{"JSON mal formé", "{", "Données invalides : JSON absent ou mal formé"},
		{"champ requis", `{"email":"a@b.c"}`, "Données invalides : title (required)"},
		{"plusieurs champs", `{"email":"pas-un-email"}`, "Données invalides : title (required), email (email)"},
		{"mauvais type", `{"title":"x","category_id":"abc"}`, "Données invalides : type incorrect pour category_id"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			w, body := record(t, func(c *gin.Context) {
				var req bindTarget
				if err := c.ShouldBindJSON(&req); err != nil {
					InvalidBody(c, err)
					return
				}
				t.Fatal("la liaison aurait dû échouer")
			}, tc.body)
			if w.Code != http.StatusBadRequest || body["error"] != tc.want || body["code"] != "bad_request" {
				t.Fatalf("statut %d, corps %v, attendu %q", w.Code, body, tc.want)
			}
			if msg := body["error"].(string); strings.Contains(msg, "bindTarget") || strings.Contains(msg, "Go struct") {
				t.Errorf("le message expose un type Go : %q", msg)
			}
		})
	}
}

func TestToSnake(t *testing.T) {
	for in, want := range map[string]string{
		"Title": "title", "CategoryID": "category_id", "ID": "id", "URL": "url", "OwnerGroupID": "owner_group_id",
	} {
		if got := toSnake(in); got != want {
			t.Errorf("toSnake(%q) = %q, attendu %q", in, got, want)
		}
	}
}
