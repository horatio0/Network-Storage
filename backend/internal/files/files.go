package files

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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

type FileInfo struct {
	Name    string    `json:"name"`
	IsDir   bool      `json:"isDir"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"modTime"`
}

func resolveSafePath(mountPath, reqPath string) (string, error) {
	cleanPath := filepath.Clean("/" + reqPath)
	target := filepath.Join(mountPath, cleanPath)
	if !strings.HasPrefix(target, filepath.Clean(mountPath)) {
		return "", os.ErrPermission
	}
	return target, nil
}

func readDirAsJSON(targetPath string) ([]FileInfo, error) {
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		return nil, err
	}
	result := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		result = append(result, FileInfo{
			Name:    entry.Name(),
			IsDir:   entry.IsDir(),
			Size:    info.Size(),
			ModTime: info.ModTime(),
		})
	}
	return result, nil
}

func ListHandler(mountPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqPath := c.Query("path")
		targetPath, err := resolveSafePath(mountPath, reqPath)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid path"})
			return
		}
		files, err := readDirAsJSON(targetPath)
		if err != nil {
			if os.IsNotExist(err) {
				c.JSON(http.StatusNotFound, gin.H{"error": "directory not found"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read directory"})
			return
		}
		c.JSON(http.StatusOK, files)
	}
}

func MkdirHandler(mountPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqPath := c.Query("path")
		if reqPath == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "path query parameter is required"})
			return
		}
		targetPath, err := resolveSafePath(mountPath, reqPath)
		if err != nil {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid path"})
			return
		}
		if err := os.MkdirAll(targetPath, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create directory"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "directory created successfully"})
	}
}

func DeleteHandler(mountPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqPath := c.Query("path")
		if reqPath == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "path query parameter is required"})
			return
		}
		targetPath, err := resolveSafePath(mountPath, reqPath)
		if err != nil || targetPath == filepath.Clean(mountPath) {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid or forbidden path"})
			return
		}
		if err := os.RemoveAll(targetPath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "deleted successfully"})
	}
}
