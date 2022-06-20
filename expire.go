package expire

import (
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/dustin/go-humanize"
)

// For all configs, a zero-value is the default.
// If any config parameter defaults to something other than a zero value,
// it will accept nil
// i.e. if dry run was true by default the DryRunConfig would take a *boolean
// but since it's not you can just put the zero value of false

const defaultFileName = ".expirations"

func DefaultDurationString() string {
	envValue := os.Getenv("EXPIRE_DEFAULT_DURATION")
	if envValue != "" {
		return envValue
	}
	return "10m"
}

func ParseDurationString(str string) (time.Duration, error) {
	return time.ParseDuration(str)
}

func DefaultDuration() time.Duration {
	var duration time.Duration

	envValue := os.Getenv("EXPIRE_DEFAULT_DURATION")
	if envValue != "" {
		var err error
		duration, err = time.ParseDuration(envValue)
		if err != nil {
			log.Printf("Error parsing EXPIRE_DEFAULT_DURATION environment variable: %s. Error: %s. Proceding with default duration", envValue, err.Error())
		}
	}
	if duration == 0 {
		return time.Minute * 10
	}
	return duration
}

type ExpirationRecords []*ExpirationRecord

func (r *ExpirationRecords) insert(rec *ExpirationRecord) {
	*r = append(*r, rec)
	sort.Sort(*r)
}

// Returns a copy of the first matching record
func (r *ExpirationRecords) getFirst(predicate func(ExpirationRecord) bool) (ExpirationRecord, bool) {
	sort.Sort(r)
	for _, rec := range *r {
		if predicate(*rec) {
			return *rec, true
		}
	}
	return ExpirationRecord{}, false
}

// Returns a copy of the first matching record
func (r *ExpirationRecords) filter(expired bool, limit int, isDelete bool, predicate func(ExpirationRecord) bool) []*ExpirationRecord {
	sort.Sort(r)
	recs := make([]*ExpirationRecord, 0)
	now := time.Now()
	deleteIdxs := make([]int, 0)
	for i, rec := range *r {
		if expired && rec.Expires.After(now) {
			break
		}
		if limit > 0 && limit <= len(recs) {
			break
		}
		if predicate(*rec) {
			deleteIdxs = append(deleteIdxs, i)
			recs = append(recs, &(*rec)) // copy it
		}
	}
	if isDelete {
		newRecs := make([]*ExpirationRecord, 0)
		for i, rec := range *r {
			if len(deleteIdxs) > 0 && i == deleteIdxs[0] {
				deleteIdxs = deleteIdxs[1:]
			} else {
				newRecs = append(newRecs, rec)
			}
		}
		*r = newRecs
		sort.Sort(r)
	}
	return recs
}

func (r *ExpirationRecords) deleteFirst(predicate func(ExpirationRecord) bool) (ExpirationRecord, bool) {
	sort.Sort(r)
	deleteIdx := -1
	for idx, rec := range *r {
		if predicate(*rec) {
			deleteIdx = idx
			break
		}
	}
	if deleteIdx == -1 {
		return ExpirationRecord{}, false
	}
	rec := *(*r)[deleteIdx]
	*r = append((*r)[:deleteIdx], (*r)[deleteIdx+1:]...)
	return rec, true
}

func (r *ExpirationRecords) updateFirst(predicate func(ExpirationRecord) bool, action func(*ExpirationRecord)) bool {
	sort.Sort(r)
	found := false
	for _, rec := range *r {
		if predicate(*rec) {
			found = true
			action(rec)
			break
		}
	}
	sort.Sort(r)
	return found
}

func (r ExpirationRecords) Len() int           { return len(r) }
func (r ExpirationRecords) Swap(i, j int)      { r[i], r[j] = r[j], r[i] }
func (r ExpirationRecords) Less(i, j int) bool { return r[i].Expires.Before(r[j].Expires) }

type ExpirationRecord struct {
	Target       string
	Expires      time.Time
	Duration     time.Duration
	ResetOnTouch bool

	// optional values
	targetFilePathAbs string
}

func (r ExpirationRecord) TargetContextual() string {
	if r.targetFilePathAbs != "" {
		wd, err := os.Getwd()
		if err != nil {
			return r.Target
		}

		relPath, err := filepath.Rel(wd, r.targetFilePathAbs)
		if err != nil {
			return r.Target
		}

		return relPath
	} else {
		return r.Target
	}
}

func (r ExpirationRecord) ExpirationRelative() string {
	return humanize.Time(r.Expires)
}

type GlobalConfig struct {
	// The name of the file instead of "expirations"
	Name string
}

func (gc GlobalConfig) getFileName() string {
	if gc.Name == "" {
		return defaultFileName
	} else {
		return gc.Name
	}
}

type DryRunConfig struct {
	IsDryRun bool
}

type BatchRunConfig struct {
	IsBatchRun bool
}

type TargetConfig struct {
	Target  string
	Targets []string
}
