package expire

import (
	"errors"
	"log"
	"os"
	"path/filepath"
	"regexp"

	"github.com/gobwas/glob"
)

type NextConfig struct {
	GlobalConfig
	Limit      int      // Match no more than this many records
	Expired    bool     // Match expired records only
	Delete     bool     // Delete the matched records
	Exist      bool     // Match records corresponding to files that exist
	NoExist    bool     // Match records corresponding to files that don't exist
	MatchGlob  []string // Match according to glob patterns
	MatchRegex []string // Match according to regex patterns
}

func Next(config *NextConfig) ([]*ExpirationRecord, error) {

	if config.Exist && config.NoExist {
		log.Println("Competing configs set, exist and noexist. Using exist.")
		config.NoExist = false
	}

	expirationsPath := getExpirationsFilePath(config.GlobalConfig)
	if expirationsPath == "" {
		return nil, errors.New("No expirations file")
	}

	records, err := readRecordsFromFile(expirationsPath)
	if err != nil {
		return nil, err
	}

	targetToFile := make(map[string]string)

	filtered := records.filter(config.Expired, config.Limit, config.Delete, func(r ExpirationRecord) bool {
		match := true

		var (
			fileRelToCurrent string
			fileExists       bool
		)
		absBase, err := filepath.Abs(filepath.Join(expirationsPath, ".."))
		if err == nil {
			relToBase := filepath.Join(absBase, r.Target)
			wd, err := os.Getwd()
			if err != nil {
				// idk
				log.Println("Couldn't get current directory")
				return true
			}
			fileRelToCurrent, err := filepath.Rel(wd, relToBase)
			if err == nil {
				fileExists = exists(fileRelToCurrent)
				if fileExists {
					targetToFile[r.Target] = relToBase
				}
			}
		}

		if config.Exist && !fileExists {
			return false
		}

		if config.NoExist && fileExists {
			return false
		}

		if config.MatchGlob != nil {
			for _, globMatch := range config.MatchGlob {
				isMatch := glob.MustCompile(globMatch).Match(fileRelToCurrent)
				if err != nil {
					continue
				}
				if !isMatch {
					match = false
				}
			}
		}
		if config.MatchRegex != nil {
			for _, regexStr := range config.MatchRegex {
				regex, err := regexp.Compile(regexStr)
				if err != nil {
					continue
				}
				if !regex.MatchString(r.Target) {
					match = false
				}
			}
		}
		return match
	})

	for _, rec := range filtered {
		if val, pres := targetToFile[rec.Target]; pres {
			rec.targetFilePathAbs = val
		}
	}

	if config.Delete {
		return filtered, writeRecordsToFile(expirationsPath, records)
	} else {
		return filtered, nil
	}

}
