package cache

import (
	"bytes"
	"fmt"
	"net/http"

	"gin-api-1/internal/auth"

	"github.com/gin-gonic/gin"
)

const (
	versionKeyPrefix = "cache:v1:version:"
	// dataKeyPrefix is the prefix for every cached read-response.
	dataKeyPrefix = "cache:v1:data:"
)

type responseRecorder struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func newResponseRecorder(w gin.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		body:           &bytes.Buffer{},
		status:         http.StatusOK,
	}
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body.Write(b)
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) WriteString(s string) (int, error) {
	r.body.WriteString(s)
	return r.ResponseWriter.WriteString(s)
}

func (r *responseRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

// ResponseCacheMiddleware caches successful GET responses in Redis and
// invalidates related entries (via version keys) after successful writes.
func ResponseCacheMiddleware(cache *RedisCache) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != http.MethodGet {
			c.Next()
			invalidateAfterWrite(c, cache)
			return
		}

		handleGet(c, cache)
	}
}

func handleGet(c *gin.Context, cache *RedisCache) {
	namespace, ok := readNamespace(c)
	if !ok {
		c.Next()
		return
	}

	version, err := cache.GetInt(c, versionKey(namespace))
	if err != nil {
		version = 0
	}

	key := fmt.Sprintf("%s%s:v%d:%s", dataKeyPrefix, namespace, version, c.Request.URL.RawQuery)

	cachedBody, err := cache.GetBytes(c, key)
	if err == nil && len(cachedBody) > 0 {
		c.Data(http.StatusOK, "application/json", cachedBody)
		c.Abort()
		return
	}

	writer := newResponseRecorder(c.Writer)
	c.Writer = writer

	c.Next()

	if writer.status == http.StatusOK && writer.body.Len() > 0 {
		cache.SetBytes(c, key, writer.body.Bytes())
	}
}

// invalidateAfterWrite bumps the version of every namespace affected by a
// successful write so previously cached responses become stale.
func invalidateAfterWrite(c *gin.Context, cache *RedisCache) {
	status := c.Writer.Status()
	if status < http.StatusOK || status >= http.StatusMultipleChoices {
		return
	}

	for _, namespace := range writeNamespaces(c) {
		cache.Increment(c, versionKey(namespace))
	}
}

// readNamespace maps a GET route to the cache namespace it belongs to.
// Only reads that hit the database and change infrequently are cached.
func readNamespace(c *gin.Context) (string, bool) {
	switch c.FullPath() {
	case "/api/v1/workspaces":
		return "workspaces:" + currentUserID(c), true
	case "/api/v1/workspaces/:id":
		return "workspace:" + c.Param("id"), true
	case "/api/v1/subscription":
		return "user-subscription:" + currentUserID(c), true
	case "/api/v1/workspaces/:id/members":
		return "workspace-members:" + c.Param("id"), true
	case "/api/v1/workspaces/:id/projects":
		return "workspace-projects:" + c.Param("id"), true
	case "/api/v1/projects/:projectID":
		return "project:" + c.Param("projectID"), true
	case "/api/v1/projects/:projectID/integrations/github":
		return "project-integration:" + c.Param("projectID"), true
	case "/api/v1/projects/:projectID/tasks":
		return "project-tasks:" + c.Param("projectID"), true
	case "/api/v1/tasks/:id":
		return "task:" + c.Param("id"), true
	default:
		return "", false
	}
}

// writeNamespaces returns the cache namespaces to invalidate for a successful
// write on the given route.
func writeNamespaces(c *gin.Context) []string {
	switch c.FullPath() {
	case "/api/v1/workspaces":
		return []string{"workspaces:" + currentUserID(c)}
	case "/api/v1/workspaces/:id":
		return []string{"workspaces:" + currentUserID(c), "workspace:" + c.Param("id")}
	case "/api/v1/checkout":
		return []string{"user-subscription:" + currentUserID(c)}
	case "/api/v1/workspaces/:id/members":
		return []string{"workspace-members:" + c.Param("id")}
	case "/api/v1/workspaces/:id/members/:userId":
		return []string{"workspace-members:" + c.Param("id")}
	case "/api/v1/workspaces/:id/projects":
		return []string{"workspace-projects:" + c.Param("id")}
	case "/api/v1/projects/:projectID":
		return []string{"project:" + c.Param("projectID")}
	case "/api/v1/projects/:projectID/integrations/github":
		return []string{"project-integration:" + c.Param("projectID")}
	case "/api/v1/projects/:projectID/integrations/github/regenerate-secret":
		return []string{"project-integration:" + c.Param("projectID")}
	case "/api/v1/projects/:projectID/tasks":
		return []string{"project-tasks:" + c.Param("projectID")}
	case "/api/v1/tasks/:id":
		return []string{"task:" + c.Param("id")}
	default:
		return nil
	}
}

func currentUserID(c *gin.Context) string {
	user, exists := c.Get("user")
	if !exists {
		return ""
	}

	userResponse, ok := user.(auth.UserResponse)
	if !ok {
		return ""
	}

	return userResponse.ID
}

func versionKey(namespace string) string {
	return versionKeyPrefix + namespace
}
