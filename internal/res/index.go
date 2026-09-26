// Package res fetches files of the EVE client from EVE's CDN.
package res

import (
	"bufio"
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/EVEShipFit/images/internal/atomicfile"
)

const (
	binariesURL  = "https://binaries.eveonline.com"
	resourcesURL = "https://resources.eveonline.com"
	userAgent    = "eveshipfit-images (+https://github.com/EVEShipFit/images)"
)

var client = &http.Client{Timeout: 5 * time.Minute}

type Entry struct {
	Path string
	MD5  string
}

// Index is keyed on the lowercased resource path, like "res:/ui/texture/icons/bpo.png".
type Index map[string]Entry

// Load keeps the index in dir, so it is only fetched once per build.
func Load(dir string, build int32) (Index, error) {
	filename := filepath.Join(dir, fmt.Sprintf("resfileindex-%d.txt", build))

	raw, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		raw, err = fetchIndex(build)
		if err != nil {
			return nil, err
		}
		if err := atomicfile.Write(filename, raw); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	return Parse(bytes.NewReader(raw))
}

func fetchIndex(build int32) ([]byte, error) {
	manifest, err := get(fmt.Sprintf("%s/eveonline_%d.txt", binariesURL, build))
	if err != nil {
		return nil, err
	}

	// The index of all resources is one of the files in the client's manifest.
	for line := range strings.SplitSeq(string(manifest), "\n") {
		fields := strings.Split(strings.TrimSpace(line), ",")
		if len(fields) >= 2 && fields[0] == "app:/resfileindex.txt" {
			return get(fmt.Sprintf("%s/%s", binariesURL, fields[1]))
		}
	}
	return nil, fmt.Errorf("client build %d lists no resource index", build)
}

// Parse reads lines of "<resource>,<cdn path>,<md5>,<size>,<compressed size>".
func Parse(r io.Reader) (Index, error) {
	index := Index{}

	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		fields := strings.Split(line, ",")
		if len(fields) < 3 {
			return nil, fmt.Errorf("resource index: malformed line %q", line)
		}
		index[strings.ToLower(fields[0])] = Entry{Path: fields[1], MD5: fields[2]}
	}
	return index, scanner.Err()
}

func (i Index) Fetch(resource string) ([]byte, error) {
	entry, ok := i[resource]
	if !ok {
		return nil, fmt.Errorf("%s: not in the resource index", resource)
	}

	raw, err := get(fmt.Sprintf("%s/%s", resourcesURL, entry.Path))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", resource, err)
	}

	sum := md5.Sum(raw)
	if hex.EncodeToString(sum[:]) != entry.MD5 {
		return nil, fmt.Errorf("%s: checksum mismatch", resource)
	}
	return raw, nil
}

type statusError struct {
	url    string
	status string
	code   int
}

func (e *statusError) Error() string {
	return fmt.Sprintf("%s: %s", e.url, e.status)
}

// A build fetches thousands of files, and the odd one fails; so retry, unless
// the CDN made clear that trying again will not help.
func get(url string) ([]byte, error) {
	var err error
	for attempt := range 3 {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * time.Second)
		}

		var raw []byte
		raw, err = getOnce(url)
		if err == nil {
			return raw, nil
		}
		if status, ok := errors.AsType[*statusError](err); ok && status.code < 500 {
			return nil, err
		}
	}
	return nil, err
}

func getOnce(url string) ([]byte, error) {
	request, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", userAgent)

	resp, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &statusError{url: url, status: resp.Status, code: resp.StatusCode}
	}
	return io.ReadAll(resp.Body)
}
