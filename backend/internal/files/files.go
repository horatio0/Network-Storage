package files

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func UploadHandler(mountPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
			return
		}

		filename := filepath.Base(file.Filename)
		if filename == "." || filename == "/" || filename == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
			return
		}

		dst := filepath.Join(mountPath, filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("upload failed: %v", err)})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  "file uploaded successfully",
			"filename": filename,
		})
	}
}

func DownloadHandler(mountPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		queryFilename := c.Query("filename")
		if queryFilename == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "filename query parameter is required"})
			return
		}

		safeFilename := filepath.Base(queryFilename)
		if safeFilename == "." || safeFilename == "/" || safeFilename == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
			return
		}

		targetPath := filepath.Join(mountPath, safeFilename)

		if _, err := os.Stat(targetPath); os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}

		c.File(targetPath)
	}
}
