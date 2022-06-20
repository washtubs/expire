package expire

import (
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/gobwas/glob"
	"github.com/pkg/errors"
)

type ScanConfig struct {
	GlobalConfig
	ForceRecursive bool
	Exclude        []string
}

func createDirectoryMatcher(config ScanConfig) (func(name string) bool, error) {
	globs := make([]glob.Glob, 0, len(config.Exclude))
	for _, excludeStr := range config.Exclude {
		g, err := glob.Compile(excludeStr)
		if err != nil {
			return nil, errors.Wrapf(err, "Error parsing %s", excludeStr)
		}
		globs = append(globs, g)
	}
	return func(name string) bool {
		for _, g := range globs {
			if g.Match(name) {
				return false
			}
		}
		return true
	}, nil
}

func doExpiredAction(config ScanConfig, info os.FileInfo, record *ExpirationRecord) {
	fmt.Printf("%s\t%s", info.Name(), record.Target)
}

func scan(config ScanConfig, info os.FileInfo) error {
	records, err := readRecordsFromFile(info.Name())
	if err != nil {
		return err
	}
	for _, record := range records {
		if record.Expires.Before(time.Now()) {
			doExpiredAction(config, info, record)
		}
	}
	return nil
}

func Scan(config ScanConfig) error {
	matcher, err := createDirectoryMatcher(config)
	if err != nil {
		return err
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working dir: %s", err)
	}

	e := filepath.Walk(cwd, func(p string, info os.FileInfo, err error) error {

		if info.IsDir() {
			if !matcher(p) {
				return filepath.SkipDir
			}
			return nil
		}

		if path.Base(info.Name()) == config.getFileName() {
			err := scan(config, info)
			if err != nil {
				log.Printf("Failed to scan %s: %s", info.Name(), err.Error())
			}
		}

		return nil
	})

	if err != nil {
		return errors.Wrap(e, "Failed to traverse directory")
	}
	return nil

}
