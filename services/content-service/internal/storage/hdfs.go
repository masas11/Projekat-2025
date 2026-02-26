package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// HDFSClient handles HDFS operations via WebHDFS REST API
type HDFSClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewHDFSClient creates a new HDFS client
func NewHDFSClient(namenodeURL string) *HDFSClient {
	if namenodeURL == "" {
		namenodeURL = "http://hdfs-namenode:9870"
	}
	return &HDFSClient{
		baseURL: namenodeURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second, // Increased timeout for large file uploads
		},
	}
}

// UploadFile uploads a file to HDFS
func (c *HDFSClient) UploadFile(localPath, hdfsPath string) error {
	// Open local file
	file, err := os.Open(localPath)
	if err != nil {
		return fmt.Errorf("failed to open local file: %w", err)
	}
	defer file.Close()

	// Create HDFS directory if it doesn't exist
	dir := filepath.Dir(hdfsPath)
	if dir != "." && dir != "/" {
		if err := c.Mkdir(dir, true); err != nil {
			// Ignore error if directory already exists
			if !strings.Contains(err.Error(), "already exists") {
				return fmt.Errorf("failed to create directory: %w", err)
			}
		}
	}

	// Step 1: Create file (redirect)
	createURL := fmt.Sprintf("%s/webhdfs/v1%s?op=CREATE&overwrite=true", c.baseURL, hdfsPath)
	req, err := http.NewRequest("PUT", createURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create file in HDFS: %w", err)
	}
	resp.Body.Close()

	// Check if we got a redirect (307)
	if resp.StatusCode == http.StatusTemporaryRedirect {
		redirectURL := resp.Header.Get("Location")
		if redirectURL == "" {
			return fmt.Errorf("no redirect location in response")
		}

		// Step 2: Upload file data to redirect URL
		fileData, err := io.ReadAll(file)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		req, err = http.NewRequest("PUT", redirectURL, bytes.NewReader(fileData))
		if err != nil {
			return fmt.Errorf("failed to create upload request: %w", err)
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to upload file: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			return fmt.Errorf("failed to upload file: status %d, body: %s", resp.StatusCode, string(body))
		}

		return nil
	}

	// If no redirect, try direct upload
	fileData, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	req, err = http.NewRequest("PUT", createURL, bytes.NewReader(fileData))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}

	resp, err = c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to upload file: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// UploadData uploads data from memory to HDFS
func (c *HDFSClient) UploadData(data []byte, hdfsPath string) error {
	// Create HDFS directory if it doesn't exist
	dir := filepath.Dir(hdfsPath)
	if dir != "." && dir != "/" {
		// Try to create directory, ignore if it already exists
		if err := c.Mkdir(dir, true); err != nil {
			// Check if directory already exists
			exists, existsErr := c.FileExists(dir)
			if existsErr != nil || !exists {
				// If directory doesn't exist and we can't create it, return error
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
			// Directory exists, continue
		}
	}

	// Retry logic for upload (HDFS can be flaky)
	maxRetries := 3
	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Wait before retry (exponential backoff)
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		// Step 1: Create file (redirect)
		createURL := fmt.Sprintf("%s/webhdfs/v1%s?op=CREATE&overwrite=true", c.baseURL, hdfsPath)
		req, err := http.NewRequest("PUT", createURL, nil)
		if err != nil {
			lastErr = fmt.Errorf("failed to create request: %w", err)
			continue
		}

		// Increase timeout for large files
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		req = req.WithContext(ctx)
		resp, err := c.httpClient.Do(req)
		cancel()
		if err != nil {
			lastErr = fmt.Errorf("failed to create file in HDFS (attempt %d/%d): %w", attempt+1, maxRetries, err)
			continue
		}
		resp.Body.Close()

		// Check if we got a redirect (307)
		if resp.StatusCode == http.StatusTemporaryRedirect {
			redirectURL := resp.Header.Get("Location")
			if redirectURL == "" {
				lastErr = fmt.Errorf("no redirect location in response")
				continue
			}

			// Step 2: Upload file data to redirect URL
			uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 120*time.Second)
			req, err = http.NewRequestWithContext(uploadCtx, "PUT", redirectURL, bytes.NewReader(data))
			if err != nil {
				uploadCancel()
				lastErr = fmt.Errorf("failed to create upload request: %w", err)
				continue
			}

			resp, err = c.httpClient.Do(req)
			uploadCancel()
			if err != nil {
				lastErr = fmt.Errorf("failed to upload file (attempt %d/%d): %w", attempt+1, maxRetries, err)
				continue
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusCreated {
				body, _ := io.ReadAll(resp.Body)
				lastErr = fmt.Errorf("failed to upload file: status %d, body: %s", resp.StatusCode, string(body))
				continue
			}

			// Success!
			return nil
		}

		// If no redirect, try direct upload
		uploadCtx, uploadCancel := context.WithTimeout(context.Background(), 120*time.Second)
		req, err = http.NewRequestWithContext(uploadCtx, "PUT", createURL, bytes.NewReader(data))
		if err != nil {
			uploadCancel()
			lastErr = fmt.Errorf("failed to create upload request: %w", err)
			continue
		}

		resp, err = c.httpClient.Do(req)
		uploadCancel()
		if err != nil {
			lastErr = fmt.Errorf("failed to upload file (attempt %d/%d): %w", attempt+1, maxRetries, err)
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusCreated {
			body, _ := io.ReadAll(resp.Body)
			lastErr = fmt.Errorf("failed to upload file: status %d, body: %s", resp.StatusCode, string(body))
			continue
		}

		// Success!
		return nil
	}

	// All retries failed
	return fmt.Errorf("upload failed after %d attempts: %w", maxRetries, lastErr)
}

// DownloadFile downloads a file from HDFS
func (c *HDFSClient) DownloadFile(hdfsPath string) ([]byte, error) {
	// Step 1: Open file (redirect)
	openURL := fmt.Sprintf("%s/webhdfs/v1%s?op=OPEN", c.baseURL, hdfsPath)
	req, err := http.NewRequest("GET", openURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to open file in HDFS: %w", err)
	}
	defer resp.Body.Close()

	// Check if we got a redirect (307)
	if resp.StatusCode == http.StatusTemporaryRedirect {
		redirectURL := resp.Header.Get("Location")
		if redirectURL == "" {
			return nil, fmt.Errorf("no redirect location in response")
		}

		// Step 2: Download file data from redirect URL
		req, err = http.NewRequest("GET", redirectURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create download request: %w", err)
		}

		resp, err = c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("failed to download file: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("failed to download file: status %d, body: %s", resp.StatusCode, string(body))
		}

		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read file data: %w", err)
		}

		return data, nil
	}

	// If no redirect, try direct download
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to download file: status %d, body: %s", resp.StatusCode, string(body))
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file data: %w", err)
	}

	return data, nil
}

// FileExists checks if a file exists in HDFS
func (c *HDFSClient) FileExists(hdfsPath string) (bool, error) {
	statusURL := fmt.Sprintf("%s/webhdfs/v1%s?op=GETFILESTATUS", c.baseURL, hdfsPath)
	req, err := http.NewRequest("GET", statusURL, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to check file status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		return true, nil
	} else if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	body, _ := io.ReadAll(resp.Body)
	return false, fmt.Errorf("unexpected status: %d, body: %s", resp.StatusCode, string(body))
}

// DeleteFile deletes a file from HDFS
func (c *HDFSClient) DeleteFile(hdfsPath string) error {
	deleteURL := fmt.Sprintf("%s/webhdfs/v1%s?op=DELETE&recursive=false", c.baseURL, hdfsPath)
	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete file: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Mkdir creates a directory in HDFS
func (c *HDFSClient) Mkdir(hdfsPath string, createParent bool) error {
	mkdirURL := fmt.Sprintf("%s/webhdfs/v1%s?op=MKDIRS&permission=755", c.baseURL, hdfsPath)
	req, err := http.NewRequest("PUT", mkdirURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errorResp struct {
			RemoteException struct {
				Exception string `json:"exception"`
				Message   string `json:"message"`
			} `json:"RemoteException"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil {
			if strings.Contains(errorResp.RemoteException.Message, "already exists") {
				return fmt.Errorf("directory already exists")
			}
		}
		return fmt.Errorf("failed to create directory: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

// GetFileStatus returns file status information
func (c *HDFSClient) GetFileStatus(hdfsPath string) (map[string]interface{}, error) {
	statusURL := fmt.Sprintf("%s/webhdfs/v1%s?op=GETFILESTATUS", c.baseURL, hdfsPath)
	req, err := http.NewRequest("GET", statusURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get file status: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get file status: status %d, body: %s", resp.StatusCode, string(body))
	}

	var result struct {
		FileStatus map[string]interface{} `json:"FileStatus"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.FileStatus, nil
}
