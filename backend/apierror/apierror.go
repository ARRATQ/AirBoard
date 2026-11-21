// Package apierror écrit les réponses d'erreur de l'API dans un format unique :
//
//	{"error": "Message lisible", "code": "not_found"}
//
// `error` est affiché tel quel par le frontend ; `code` est stable et dérivé du
// statut HTTP, sauf pour quelques cas métier nommés (RespondCode).
package apierror

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"airboard/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// Codes métier : un client peut s'y fier pour réagir à un cas précis.
const (
	CodeJWTSecretChanged   = "jwt_secret_changed"
	CodeCommentsDisabled   = "comments_disabled"
	CodeCommentTooLong     = "comment_too_long"
	CodeWeakPassword       = "weak_password"
	CodeFileSecurity       = "file_security_violation"
	CodeOAuthStateMismatch = "oauth_state_mismatch"
)

// CodeForStatus renvoie le code générique d'un statut HTTP d'erreur.
func CodeForStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusForbidden:
		return "forbidden"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusRequestEntityTooLarge:
		return "payload_too_large"
	case http.StatusUnprocessableEntity:
		return "unprocessable_entity"
	case http.StatusTooManyRequests:
		return "too_many_requests"
	case http.StatusServiceUnavailable:
		return "service_unavailable"
	}
	if status >= 500 {
		return "internal_error"
	}
	return "error"
}

// Body construit le corps d'erreur standard.
func Body(status int, message string) models.ErrorResponse {
	return models.ErrorResponse{Error: message, Code: CodeForStatus(status)}
}

// Respond écrit une erreur dont le code découle du statut.
func Respond(c *gin.Context, status int, message string) {
	c.JSON(status, Body(status, message))
}

// RespondCode écrit une erreur avec un code métier explicite.
func RespondCode(c *gin.Context, status int, code, message string) {
	c.JSON(status, models.ErrorResponse{Error: message, Code: code})
}

// Abort écrit l'erreur et interrompt la chaîne de middlewares.
func Abort(c *gin.Context, status int, message string) {
	c.AbortWithStatusJSON(status, Body(status, message))
}

// InvalidBody répond 400 à un corps de requête illisible ou invalide, avec un
// message qui nomme les champs fautifs sans exposer les types Go internes.
func InvalidBody(c *gin.Context, err error) {
	Respond(c, http.StatusBadRequest, BindMessage(err))
}

// BindMessage traduit une erreur de liaison Gin (JSON, validation) en message
// lisible : « Données invalides : title (required), email (email) ».
func BindMessage(err error) string {
	const prefix = "Données invalides"
	var verrs validator.ValidationErrors
	if errors.As(err, &verrs) {
		fields := make([]string, 0, len(verrs))
		for _, fe := range verrs {
			fields = append(fields, fmt.Sprintf("%s (%s)", jsonFieldName(fe), fe.Tag()))
		}
		return prefix + " : " + strings.Join(fields, ", ")
	}
	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) && typeErr.Field != "" {
		return prefix + " : type incorrect pour " + typeErr.Field
	}
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) || errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return prefix + " : JSON absent ou mal formé"
	}
	return prefix
}

// jsonFieldName : nom du champ tel que l'envoie le client (snake_case) plutôt
// que le nom du champ Go.
func jsonFieldName(fe validator.FieldError) string {
	return toSnake(fe.Field())
}

func toSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 && !(s[i-1] >= 'A' && s[i-1] <= 'Z') {
				b.WriteByte('_')
			}
			b.WriteRune(r + ('a' - 'A'))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}
