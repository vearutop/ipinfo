// Package cloud makes an index of cloud providers.
package cloud

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/vearutop/netrie"
	"github.com/vearutop/netrie/lists"
)

// LoadCloud fills CIDRIndex with networks from
// https://github.com/disposable/cloud-ip-ranges/tree/master/txt.
func LoadCloud(tr netrie.Adder) error {
	apiURL := "https://api.github.com/repos/disposable/cloud-ip-ranges/contents/txt"

	type FileEntry struct {
		Name        string `json:"name"`
		Type        string `json:"type"`
		DownloadURL string `json:"download_url"`
	}

	// Request directory listing
	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")

	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return fmt.Errorf("listing directory: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Println("closing response body:", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API request failed: %s", resp.Status)
	}

	var entries []FileEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	for _, entry := range entries {
		if entry.Type != "file" {
			continue
		}

		name := strings.TrimSuffix(entry.Name, ".txt")

		if err := lists.LoadFromTextGroupCIDRs(entry.DownloadURL, tr, name); err != nil {
			return err
		}
	}

	return nil
}

// LoadCloudLocal fills CIDRIndex with networks from
// https://github.com/disposable/cloud-ip-ranges/tree/master/txt.
func LoadCloudLocal(tr netrie.Adder, dir string) error {
	dir = path.Join(dir, "txt")

	meta := tr.Metadata()
	if meta.Description == "" {
		meta.Description = "Cloud providers from github.com/disposable/cloud-ip-ranges"
	}

	if meta.Name == "" {
		meta.Name = "Cloud providers"
	}

	tr.Metadata().BuildDate = time.Now().UTC()

	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		// Skip directories
		if entry.IsDir() {
			continue
		}

		// Check if file has .txt extension (case-insensitive)
		if strings.HasSuffix(strings.ToLower(entry.Name()), ".txt") {
			name := strings.TrimSuffix(entry.Name(), ".txt")

			if err := lists.LoadFromTextGroupCIDRs(filepath.Join(dir, entry.Name()), tr, name); err != nil {
				return err
			}
		}
	}

	return nil
}
