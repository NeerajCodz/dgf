package github

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// LFSPointer represents a Git LFS pointer file
type LFSPointer struct {
	Version string // version https://git-lfs.github.com/spec/v1
	OID     string // oid sha256:xxx
	Size    int64  // size in bytes
}

// LFSObject represents an LFS object in the batch response
type LFSObject struct {
	OID  string `json:"oid"`
	Size int64  `json:"size"`
}

// LFSBatchRequest is the request body for LFS batch API
type LFSBatchRequest struct {
	Operation string      `json:"operation"`
	Objects   []LFSObject `json:"objects"`
	Transfers []string    `json:"transfers,omitempty"`
}

// LFSAction represents a download action in the LFS response
type LFSAction struct {
	Href      string            `json:"href"`
	Header    map[string]string `json:"header,omitempty"`
	ExpiresIn int               `json:"expires_in,omitempty"`
}

// LFSBatchObject represents an object in the batch response
type LFSBatchObject struct {
	OID     string               `json:"oid"`
	Size    int64                `json:"size"`
	Actions map[string]LFSAction `json:"actions,omitempty"`
	Error   *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// LFSBatchResponse is the response from LFS batch API
type LFSBatchResponse struct {
	Transfer string           `json:"transfer,omitempty"`
	Objects  []LFSBatchObject `json:"objects"`
}

// IsLFSPointer checks if the content is a Git LFS pointer file
func IsLFSPointer(content []byte) bool {
	// LFS pointer files start with "version https://git-lfs.github.com/spec/v1"
	return bytes.HasPrefix(content, []byte("version https://git-lfs.github.com/spec/v1"))
}

// ParseLFSPointer parses an LFS pointer file content
func ParseLFSPointer(content []byte) (*LFSPointer, error) {
	lines := strings.Split(string(content), "\n")
	pointer := &LFSPointer{}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "version ") {
			pointer.Version = strings.TrimPrefix(line, "version ")
		} else if strings.HasPrefix(line, "oid sha256:") {
			pointer.OID = strings.TrimPrefix(line, "oid ")
		} else if strings.HasPrefix(line, "size ") {
			sizeStr := strings.TrimPrefix(line, "size ")
			size, err := strconv.ParseInt(sizeStr, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid size in LFS pointer: %v", err)
			}
			pointer.Size = size
		}
	}

	if pointer.OID == "" {
		return nil, fmt.Errorf("missing OID in LFS pointer")
	}

	return pointer, nil
}

// GetLFSDownloadURL gets the actual download URL for an LFS object
func (c *Client) GetLFSDownloadURL(owner, repo string, pointer *LFSPointer) (string, map[string]string, error) {
	// Construct LFS batch API URL
	lfsURL := fmt.Sprintf("https://github.com/%s/%s.git/info/lfs/objects/batch", owner, repo)

	// Create batch request
	oidParts := strings.SplitN(pointer.OID, ":", 2)
	if len(oidParts) != 2 {
		return "", nil, fmt.Errorf("invalid OID format: %s", pointer.OID)
	}
	oid := oidParts[1] // Remove sha256: prefix

	batchReq := LFSBatchRequest{
		Operation: "download",
		Objects: []LFSObject{
			{OID: oid, Size: pointer.Size},
		},
		Transfers: []string{"basic"},
	}

	body, err := json.Marshal(batchReq)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal batch request: %v", err)
	}

	req, err := http.NewRequest("POST", lfsURL, bytes.NewReader(body))
	if err != nil {
		return "", nil, fmt.Errorf("failed to create LFS request: %v", err)
	}

	req.Header.Set("Content-Type", "application/vnd.git-lfs+json")
	req.Header.Set("Accept", "application/vnd.git-lfs+json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("LFS batch request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("LFS batch API returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var batchResp LFSBatchResponse
	if err := json.NewDecoder(resp.Body).Decode(&batchResp); err != nil {
		return "", nil, fmt.Errorf("failed to parse LFS batch response: %v", err)
	}

	if len(batchResp.Objects) == 0 {
		return "", nil, fmt.Errorf("no objects in LFS batch response")
	}

	obj := batchResp.Objects[0]
	if obj.Error != nil {
		return "", nil, fmt.Errorf("LFS error: %s", obj.Error.Message)
	}

	downloadAction, ok := obj.Actions["download"]
	if !ok {
		return "", nil, fmt.Errorf("no download action in LFS response")
	}

	return downloadAction.Href, downloadAction.Header, nil
}

// FetchLFSFile downloads a file through Git LFS
func (c *Client) FetchLFSFile(owner, repo string, pointer *LFSPointer) ([]byte, error) {
	downloadURL, headers, err := c.GetLFSDownloadURL(owner, repo, pointer)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %v", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("LFS download failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("LFS download returned %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
