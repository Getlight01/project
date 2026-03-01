package storage

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/disintegration/imaging"
)

type FileStorage struct {
	basePath string
}

func NewFileStorage(basePath string) *FileStorage {
	return &FileStorage{
		basePath: basePath,
	}
}

func (fs *FileStorage) SaveUpload(file io.Reader, filename string) (string, string, error) {

	ext := filepath.Ext(filename)
	if ext == "" {
		ext = ".jpg"
	}

	timestamp := time.Now().UnixNano()
	uniqueName := fmt.Sprintf("%d%s", timestamp, ext)

	uploadDir := filepath.Join(fs.basePath, "uploads")
	thumbDir := filepath.Join(fs.basePath, "uploads", "thumbnails")

	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create upload directory: %w", err)
	}
	if err := os.MkdirAll(thumbDir, 0755); err != nil {
		return "", "", fmt.Errorf("failed to create thumbnail directory: %w", err)
	}

	originalPath := filepath.Join(uploadDir, uniqueName)
	originalFile, err := os.Create(originalPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to create file: %w", err)
	}
	defer originalFile.Close()

	if _, err := io.Copy(originalFile, file); err != nil {
		return "", "", fmt.Errorf("failed to save file: %w", err)
	}

	thumbPath := filepath.Join(thumbDir, "thumb_"+uniqueName)
	if err := fs.generateThumbnail(originalPath, thumbPath); err != nil {

		fmt.Printf("Failed to generate thumbnail: %v\n", err)
		thumbPath = originalPath
	}

	originalURL := "/uploads/" + uniqueName
	thumbURL := "/uploads/thumbnails/thumb_" + uniqueName

	return originalURL, thumbURL, nil
}

func (fs *FileStorage) DownloadFromURL(url string) (string, string, error) {

	resp, err := http.Get(url)
	if err != nil {
		return "", "", fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("bad status: %s", resp.Status)
	}

	filename := "downloaded.jpg"
	if cd := resp.Header.Get("Content-Disposition"); cd != "" {
		if parts := strings.Split(cd, "filename="); len(parts) > 1 {
			filename = strings.Trim(parts[1], `"`)
		}
	}

	return fs.SaveUpload(resp.Body, filename)
}

func (fs *FileStorage) generateThumbnail(sourcePath, destPath string) error {

	src, err := imaging.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to open image: %w", err)
	}

	thumb := imaging.Fit(src, 300, 300, imaging.Lanczos)

	if err := imaging.Save(thumb, destPath); err != nil {
		return fmt.Errorf("failed to save thumbnail: %w", err)
	}

	return nil
}

func (fs *FileStorage) DeleteFile(url string) error {
	if url == "" {
		return nil
	}

	filePath := filepath.Join(fs.basePath, url)

	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (fs *FileStorage) GetFilePath(url string) string {
	return filepath.Join(fs.basePath, url)
}

func (fs *FileStorage) GetFullPath(url string) string {
	return fs.GetFilePath(url)
}

func (fs *FileStorage) GetBasePath() string {
	return fs.basePath
}
