package files

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func UploadHandler(mountPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reader, err := c.Request.MultipartReader()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read multipart"})
			return
		}

		for {
			part, err := reader.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read part"})
				return
			}
			if part.FormName() == "file" {
				filename := filepath.Base(part.FileName())
				if filename == "." || filename == "/" || filename == "" {
					c.JSON(http.StatusBadRequest, gin.H{"error": "invalid filename"})
					return
				}

				dst := filepath.Join(mountPath, filename)
				out, err := os.Create(dst)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create file"})
					return
				}
				defer out.Close()

				if _, err := io.Copy(out, part); err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
					return
				}

				c.JSON(http.StatusOK, gin.H{
					"message":  "file uploaded successfully",
					"filename": filename,
				})
				return
			}
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "no file provided"})
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
