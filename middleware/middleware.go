package middleware

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/Jojojojodr/bonfirec2"
	"github.com/gin-gonic/gin"
)

var allowedUploadExtensions = map[string]struct{}{
	".txt":  {},
	".log":  {},
	".json": {},
	".csv":  {},
}

const maxUploadFileSize int64 = 10 << 20

func LoggingMiddleware(c *gin.Context) {
	start := time.Now()

	c.Next()

	latency := time.Since(start)
	statusCode := c.Writer.Status()
	method := c.Request.Method
	path := c.Request.URL.Path
	query := c.Request.URL.RawQuery
	userAgent := c.Request.UserAgent()

	if query != "" {
		path = path + "?" + query
	}

	requestDetails := fmt.Sprintf("%s | status=%d | latency=%s | ua=%q", path, statusCode, latency, userAgent)
	err := bonfirec2.LogRequest(method, requestDetails)
	if err != nil {
		fmt.Printf("Failed to log request: %v\n", err)
	}
}

func ValidateUploadFile(c *gin.Context) {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		c.Abort()
		return
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	if _, ok := allowedUploadExtensions[ext]; !ok {
		c.JSON(http.StatusUnsupportedMediaType, gin.H{
			"error":              "unsupported file type",
			"allowed_extensions": []string{".txt", ".log", ".json", ".csv"},
		})
		c.Abort()
		return
	}

	if fileHeader.Size <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "uploaded file is empty"})
		c.Abort()
		return
	}

	if fileHeader.Size > maxUploadFileSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "file exceeds 10MB limit"})
		c.Abort()
		return
	}

	c.Next()
}
