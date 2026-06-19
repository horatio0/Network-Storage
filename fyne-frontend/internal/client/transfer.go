package client

import (
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

func UploadFile(c *HTTPClient, ip, port, filePath, targetPath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	pr, pw := io.Pipe()
	w := multipart.NewWriter(pw)
	go streamUpload(file, pw, w, filePath)
	return sendUploadReq(c, pr, w.FormDataContentType(), ip, port, targetPath)
}

func streamUpload(file *os.File, pw *io.PipeWriter, w *multipart.Writer, filePath string) {
	defer file.Close()
	defer pw.Close()
	part, err := w.CreateFormFile("file", filepath.Base(filePath))
	if err == nil {
		io.Copy(part, file)
	}
	w.Close()
}

func sendUploadReq(c *HTTPClient, pr io.Reader, ct, ip, port, target string) error {
	u := fmt.Sprintf("http://%s:%s/api/v1/files/upload?path=%s", ip, port, url.QueryEscape(target))
	req, _ := http.NewRequest("POST", u, pr)
	req.Header.Set("Content-Type", ct)
	_, err := c.DoRequest(req)
	return err
}

func DownloadFile(c *HTTPClient, ip, port, path, savePath string) error {
	u := fmt.Sprintf("http://%s:%s/api/v1/files/download?path=%s", ip, port, url.QueryEscape(path))
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
