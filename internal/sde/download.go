package sde

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/EVEShipFit/images/internal/atomicfile"
)

const (
	latestURL   = "https://developers.eveonline.com/static-data/tranquility/latest.jsonl"
	downloadURL = "https://developers.eveonline.com/static-data/tranquility/eve-online-static-data-%d-jsonl.zip"
)

var client = &http.Client{Timeout: 30 * time.Minute}

// LatestBuild returns the build number of the SDE currently on Tranquility.
func LatestBuild() (int32, error) {
	resp, err := client.Get(latestURL)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("%s: %s", latestURL, resp.Status)
	}

	decoder := json.NewDecoder(resp.Body)
	for {
		var record struct {
			Key         string `json:"_key"`
			BuildNumber int32  `json:"buildNumber"`
		}
		if err := decoder.Decode(&record); err == io.EOF {
			break
		} else if err != nil {
			return 0, err
		}

		if record.Key == "sde" {
			return record.BuildNumber, nil
		}
	}
	return 0, fmt.Errorf("%s: no record with key 'sde'", latestURL)
}

// Download fetches the JSONL SDE of the given build into dir, and returns the
// path to it. An already downloaded build is not fetched again.
func Download(dir string, build int32) (string, error) {
	filename := filepath.Join(dir, fmt.Sprintf("eve-online-static-data-%d-jsonl.zip", build))

	if _, err := os.Stat(filename); err == nil {
		return filename, nil
	}

	url := fmt.Sprintf(downloadURL, build)
	resp, err := client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("%s: %s", url, resp.Status)
	}

	if err := atomicfile.Copy(filename, resp.Body); err != nil {
		return "", err
	}
	return filename, nil
}

// LocalBuild returns the build number of the newest SDE in dir, or zero if
// there is none.
func LocalBuild(dir string) int32 {
	matches, err := filepath.Glob(filepath.Join(dir, "eve-online-static-data-*-jsonl.zip"))
	if err != nil {
		return 0
	}

	var newest int32
	for _, match := range matches {
		var build int32
		if _, err := fmt.Sscanf(filepath.Base(match), "eve-online-static-data-%d-jsonl.zip", &build); err != nil {
			continue
		}
		if build > newest {
			newest = build
		}
	}
	return newest
}
