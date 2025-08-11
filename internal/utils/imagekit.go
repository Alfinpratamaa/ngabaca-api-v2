package utils

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"ngabaca/config"
	"strings"
)

type ImageKitUploadResponse struct {
	URL string `json:"url"`
}

type ImageKitFile struct {
	FileID   string `json:"fileId"`
	FilePath string `json:"filePath"`
}

// UploadToImageKit mengupload file ke ImageKit dan mengembalikan URL
func UploadToImageKit(cfg config.Config, fileBytes []byte, fileName string, folder string) (string, error) {
	uploadURL := "https://upload.imagekit.io/api/v1/files/upload"

	// Encode file jadi base64
	fileBase64 := base64.StdEncoding.EncodeToString(fileBytes)

	// Buat form-data
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("file", "data:image/jpeg;base64,"+fileBase64)
	_ = writer.WriteField("fileName", fileName)

	// Kalau folder dikasih, tambahkan
	if folder != "" {
		_ = writer.WriteField("folder", folder)
	}

	writer.Close()

	// Request
	req, err := http.NewRequest("POST", uploadURL, body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetBasicAuth(cfg.ImageKitPrivateKey, "")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	respBytes, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return "", fmt.Errorf("imagekit upload failed: %s", string(respBytes))
	}

	var uploadResp ImageKitUploadResponse
	if err := json.Unmarshal(respBytes, &uploadResp); err != nil {
		return "", err
	}

	return uploadResp.URL, nil
}

func DeleteFromImageKit(cfg config.Config, fileURL string) error {
	if fileURL == "" {
		return errors.New("file URL is empty")
	}

	// Ambil filePath dari URL (hapus domain)
	parsedURL, err := url.Parse(fileURL)
	if err != nil {
		return err
	}
	filePath := strings.TrimPrefix(parsedURL.Path, "/")
	filePath = strings.TrimPrefix(filePath, cfg.ImageKitID+"/") // hapus imagekit id kalau ada

	// 1. Cari file di ImageKit untuk dapatkan fileId
	listURL := fmt.Sprintf("https://api.imagekit.io/v1/files?path=%s", url.QueryEscape(filePath))
	req, err := http.NewRequest("GET", listURL, nil)
	if err != nil {
		return err
	}
	req.SetBasicAuth(cfg.ImageKitPrivateKey, "")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to list file: %s", string(body))
	}

	var files []ImageKitFile
	if err := json.NewDecoder(res.Body).Decode(&files); err != nil {
		return err
	}
	if len(files) == 0 {
		return errors.New("file not found in ImageKit")
	}

	fileID := files[0].FileID

	// 2. Hapus file
	deleteURL := fmt.Sprintf("https://api.imagekit.io/v1/files/%s", fileID)
	delReq, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return err
	}
	delReq.SetBasicAuth(cfg.ImageKitPrivateKey, "")

	delRes, err := client.Do(delReq)
	if err != nil {
		return err
	}
	defer delRes.Body.Close()

	if delRes.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(delRes.Body)
		return fmt.Errorf("failed to delete file: %s", string(body))
	}

	return nil
}
