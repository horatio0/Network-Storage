package files

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUploadDownloadFlow(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tmpDir, err := os.MkdirTemp("", "files_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	router := gin.New()
	router.POST("/upload", UploadHandler(tmpDir))
	router.GET("/download", DownloadHandler(tmpDir))

	t.Run("upload successfully", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.txt")
		part.Write([]byte("hello world"))
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.Code)
		}
	})

	t.Run("download successfully", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/download?filename=test.txt", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		if resp.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, resp.Code)
		}

		if resp.Body.String() != "hello world" {
			t.Errorf("expected body %q, got %q", "hello world", resp.Body.String())
		}
	})

	t.Run("path traversal attempt download", func(t *testing.T) {
		// Even if they pass ../test.txt, filepath.Base will reduce it to test.txt
		// We can test that it still finds test.txt or errors appropriately if we didn't exist.
		// Let's test a non-existent file path traversal.
		req := httptest.NewRequest(http.MethodGet, "/download?filename=../../../etc/passwd", nil)
		resp := httptest.NewRecorder()

		router.ServeHTTP(resp, req)

		// filepath.Base will turn it into 'passwd', which doesn't exist in tmpDir.
		if resp.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, resp.Code)
		}
	})
}
