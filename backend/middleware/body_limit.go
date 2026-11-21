package middleware

import (
	"net/http"
	"strings"

	"airboard/apierror"

	"github.com/gin-gonic/gin"
)

const (
	// MaxJSONBodyBytes borne les requêtes classiques (JSON, formulaires).
	MaxJSONBodyBytes int64 = 5 << 20 // 5 Mo
	// MaxUploadBodyBytes borne les envois de fichiers (médias jusqu'à 10 Mo + enveloppe multipart).
	MaxUploadBodyBytes int64 = 12 << 20 // 12 Mo
)

// LimitRequestBody refuse les corps de requête trop gros : 413 immédiat si
// Content-Length dépasse la limite, sinon lecture plafonnée par http.MaxBytesReader.
func LimitRequestBody() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil || c.Request.Body == http.NoBody {
			c.Next()
			return
		}

		limit := bodyLimitFor(c.GetHeader("Content-Type"))
		if c.Request.ContentLength > limit {
			apierror.Abort(c, http.StatusRequestEntityTooLarge, "Requête trop volumineuse")
			return
		}

		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}

func bodyLimitFor(contentType string) int64 {
	if strings.HasPrefix(strings.ToLower(contentType), "multipart/") {
		return MaxUploadBodyBytes
	}
	return MaxJSONBodyBytes
}
