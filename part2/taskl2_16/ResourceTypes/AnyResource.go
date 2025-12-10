package ResourceTypes

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
)

type AnyResource interface {
	RemoteUrlString() string
	Path() string
}

func DownloadResource[T AnyResource](resource T, pathToSave url.URL) (string, error) {
	remoteUrl, err := url.Parse(resource.RemoteUrlString())
	if err != nil {
		return "", fmt.Errorf("failed to parse remote URL: %w", err)
	}
	remoteUrl = remoteUrl.JoinPath(resource.Path())

	resp, err := http.Get(remoteUrl.String())
	if err != nil {
		return "", fmt.Errorf("failed to get resource from URL: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status code: %s", resp.Status)
	}

	localPath := filepath.Join(pathToSave.String(), resource.Path())
	fileError := os.MkdirAll(filepath.Dir(localPath), os.ModePerm)
	if fileError != nil {
		return "", fmt.Errorf("failed to create directory: %w", fileError)
	}
	outputFile, err := os.Create(localPath)

	if err != nil {
		return "", fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	_, err = io.Copy(outputFile, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to copy resource data: %w", err)
	}

	return localPath, nil
}
