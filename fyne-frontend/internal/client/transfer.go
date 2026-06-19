package client

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
)

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
	url := fmt.Sprintf("http://%s:%s/api/v1/files/upload", ip, port)
	req, _ := http.NewRequest("POST", url, body)
	req.Header.Set("Content-Type", ct)
	_, err := c.DoRequest(req)
	return err
}

func DownloadFile(c *HTTPClient, ip, port, filename, savePath string) error {
	url := fmt.Sprintf("http://%s:%s/api/v1/files/download?filename=%s", ip, port, filename)
	req, _ := http.NewRequest("GET", url, nil)
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
