// Command images converts EVE's icons to WebP, and writes images.dat to look them up.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/EVEShipFit/images/internal/fbs"
	"github.com/EVEShipFit/images/internal/icons"
	"github.com/EVEShipFit/images/internal/res"
	"github.com/EVEShipFit/images/internal/sde"
)

const datFile = "images.dat"

var (
	sdeDir   = flag.String("sde-dir", "sde", "directory the SDE is downloaded to")
	cacheDir = flag.String("cache-dir", "cache", "directory the resource index and converted images are kept in")
	outDir   = flag.String("out", "dist", "directory to write to")
	buildNum = flag.Int("build", 0, "SDE and client build to use; defaults to the latest")
	jobs     = flag.Int("jobs", 2*runtime.NumCPU(), "images to fetch and convert at the same time")
)

func usage() {
	fmt.Fprint(os.Stderr, `usage: images [flags] <command>

commands:
  download   fetch the SDE and the client's resource index
  build      convert the images and write images.dat
  compare    tell whether the build differs from the one in a directory; "compare <dir>"

flags:
`)
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage

	// The command comes first, so that flags can be written after it.
	args := os.Args[1:]
	command := ""
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		command, args = args[0], args[1:]
	}
	if err := flag.CommandLine.Parse(args); err != nil {
		os.Exit(2)
	}

	if err := run(command, flag.Args()); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func run(command string, args []string) error {
	switch command {
	case "download":
		_, build, err := download()
		if err != nil {
			return err
		}
		_, err = res.Load(*cacheDir, build)
		return err

	case "build":
		data, err := load()
		if err != nil {
			return err
		}
		// The client shares its build number with the SDE.
		index, err := res.Load(*cacheDir, data.BuildNumber)
		if err != nil {
			return err
		}

		plan, err := icons.NewPlan(data, index)
		if err != nil {
			return err
		}
		if len(plan.Missing) > 0 {
			fmt.Fprintf(os.Stderr, "%d published type(s) have no icon: %v\n", len(plan.Missing), plan.Missing)
		}

		if err := os.MkdirAll(*outDir, 0o755); err != nil {
			return err
		}
		if err := icons.Build(plan, index, *cacheDir, *outDir, *jobs); err != nil {
			return err
		}
		filename := filepath.Join(*outDir, datFile)
		if err := fbs.Write(plan, filename); err != nil {
			return err
		}

		size, err := dirSize(filepath.Join(*outDir, icons.ImagesDir))
		if err != nil {
			return err
		}
		info, err := os.Stat(filename)
		if err != nil {
			return err
		}
		fmt.Printf("wrote %s (%d images, %.1f MiB)\n", *outDir, len(plan.Sources()), float64(size)/1024/1024)
		fmt.Printf("wrote %s (%.1f KiB)\n", filename, float64(info.Size())/1024)
		fmt.Printf("%d types, %d market groups, %d meta groups, %d UI textures\n",
			len(plan.Types), len(plan.MarketGroups), len(plan.MetaGroups), len(plan.UI))
		return nil

	case "compare":
		if len(args) == 0 {
			return fmt.Errorf("compare needs the directory of an earlier build")
		}
		// Images are named after a hash of their content, so a changed image changes images.dat too.
		previous, err := os.ReadFile(filepath.Join(args[0], datFile))
		if err != nil {
			return err
		}
		current, err := os.ReadFile(filepath.Join(*outDir, datFile))
		if err != nil {
			return err
		}
		if bytes.Equal(previous, current) {
			fmt.Println("unchanged")
		} else {
			fmt.Println("changed")
		}
		return nil

	default:
		usage()
		return fmt.Errorf("unknown command %q", command)
	}
}

func download() (string, int32, error) {
	build := int32(*buildNum)
	if build == 0 {
		latest, err := sde.LatestBuild()
		if err != nil {
			return "", 0, err
		}
		build = latest
	}

	filename, err := sde.Download(*sdeDir, build)
	if err != nil {
		return "", 0, err
	}

	fmt.Fprintf(os.Stderr, "using SDE build %d\n", build)
	return filename, build, nil
}

// load uses the SDE already on disk when there is one, so that building does
// not need the network.
func load() (*sde.Data, error) {
	build := int32(*buildNum)
	if build == 0 {
		build = sde.LocalBuild(*sdeDir)
	}

	if build == 0 {
		filename, downloaded, err := download()
		if err != nil {
			return nil, err
		}
		return sde.Load(filename, downloaded)
	}

	filename, err := sde.Download(*sdeDir, build)
	if err != nil {
		return nil, err
	}
	return sde.Load(filename, build)
}

func dirSize(dir string) (int64, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, err
	}
	var size int64
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return 0, err
		}
		size += info.Size()
	}
	return size, nil
}
