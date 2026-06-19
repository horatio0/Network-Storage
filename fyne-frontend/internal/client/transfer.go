package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type FileInfo struct {
	Name    string `json:"name"`
	IsDir   bool   `json:"isDir"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
}

func ListFiles(c *HTTPClient, ip, port, path string) ([]FileInfo, error) {
	u := fmt.Sprintf("http://%s:%s/api/v1/files/list?path=%s", ip, port, url.QueryEscape(path))
	req, _ := http.NewRequest("GET", u, nil)
	body, err := c.DoRequest(req)
	if err != nil {
		return nil, err
	}
	var list []FileInfo
	if err := json.Unmarshal(body, &list); err != nil {
		return nil, err
	}
	return list, nil
}

func UploadFile(c *HTTPClient, ip, port, filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, _ := w.CreateFormFile("file", filepath.Base(filePath))
	io.Copy(part, file)
	w.Close()
	return executeUpload(c, ip, port, body, w.FormDataContentType())
}

func executeUpload(c *HTTPClient, ip, port string, body io.Reader, ct string) error {
	u := fmt.Sprintf("http://%s:%s/api/v1/files/upload", ip, port)
	req, _ := http.NewRequest("POST", u, body)
	req.Header.Set("Content-Type", ct)
	_, err := c.DoRequest(req)
	return err
}

func DownloadFile(c *HTTPClient, ip, port, filename, savePath string) error {
	u := fmt.Sprintf("http://%s:%s/api/v1/files/download?filename=%s", ip, port, filename)
	req, _ := http.NewRequest("GET", u, nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return saveResponse(resp, savePath)
}

func saveResponse(resp *http.Response, savePath string) error {
	if resp.StatusCode != 200 {
		return fmt.Errorf("status: %d", resp.StatusCode)
	}
	out, err := os.Create(savePath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

func Mkdir(c *HTTPClient, ip, port, path string) error {
	u := fmt.Sprintf("http://%s:%s/api/v1/files/mkdir?path=%s", ip, port, url.QueryEscape(path))
	req, _ := http.NewRequest("POST", u, nil)
	_, err := c.DoRequest(req)
	return err
}

func DeletePath(c *HTTPClient, ip, port, path string) error {
	u := fmt.Sprintf("http://%s:%s/api/v1/files/delete?path=%s", ip, port, url.QueryEscape(path))
	req, _ := http.NewRequest("DELETE", u, nil)
	_, err := c.DoRequest(req)
	return err
}
