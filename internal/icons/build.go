package icons

import (
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"sync"

	"github.com/EVEShipFit/images/internal/atomicfile"
	"github.com/EVEShipFit/images/internal/res"
)

const ImagesDir = "images"

// Build keeps converted images in cacheDir, so only new or changed ones are fetched again.
func Build(plan *Plan, index res.Index, cacheDir, outDir string, jobs int) error {
	sources := plan.Sources()
	cache := filepath.Join(cacheDir, ImagesDir)

	if err := convertAll(sources, index, cache, jobs); err != nil {
		return err
	}

	files := map[string]bool{}
	for _, source := range sources {
		files[source.fileName()] = true
	}
	if err := prune(cache, files); err != nil {
		return err
	}

	out := filepath.Join(outDir, ImagesDir)
	if err := os.RemoveAll(out); err != nil {
		return err
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	for _, file := range slices.Sorted(maps.Keys(files)) {
		if err := copyFile(filepath.Join(cache, file), filepath.Join(out, file)); err != nil {
			return err
		}
	}
	return nil
}

func convertAll(sources []Source, index res.Index, cache string, jobs int) error {
	var (
		mutex    sync.Mutex
		errs     []error
		fetched  int
		work     = make(chan Source)
		finished sync.WaitGroup
	)

	for range jobs {
		finished.Go(func() {
			for source := range work {
				didFetch, err := ensure(source, filepath.Join(cache, source.fileName()), index)

				mutex.Lock()
				if err != nil {
					errs = append(errs, err)
				}
				if didFetch {
					fetched++
					if fetched%500 == 0 {
						fmt.Fprintf(os.Stderr, "converted %d images\n", fetched)
					}
				}
				mutex.Unlock()
			}
		})
	}

	for _, source := range sources {
		work <- source
	}
	close(work)
	finished.Wait()

	if len(errs) > 0 {
		return fmt.Errorf("%d image(s) failed, the first: %w", len(errs), errs[0])
	}
	fmt.Fprintf(os.Stderr, "%d images, %d of them newly converted\n", len(sources), fetched)
	return nil
}

func ensure(source Source, filename string, index res.Index) (bool, error) {
	if _, err := os.Stat(filename); err == nil {
		return false, nil
	}

	raw, err := index.Fetch(source.Resource)
	if err != nil {
		return false, err
	}
	converted, err := convert(raw, source.Profile)
	if err != nil {
		return false, fmt.Errorf("%s: %w", source.Resource, err)
	}
	return true, atomicfile.Write(filename, converted)
}

// prune keeps the cache from growing forever.
func prune(cache string, files map[string]bool) error {
	entries, err := os.ReadDir(cache)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if !files[entry.Name()] {
			if err := os.Remove(filepath.Join(cache, entry.Name())); err != nil {
				return err
			}
		}
	}
	return nil
}

func copyFile(from, to string) error {
	in, err := os.Open(from)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(to)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}
