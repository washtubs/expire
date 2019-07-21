package expire

import "errors"
import "time"

type CheckConfig struct {
	GlobalConfig
	TargetConfig
}

func checkCheck(config *CheckConfig) error {
	if config.Target == "" {
		return errors.New("No target")
	}
	return nil
}

type CheckResponse int

const (
	TrackedUnexpired CheckResponse = 0
	TrackedExpired                 = 1
	Untracked                      = 2
)

func Check(config *CheckConfig) (CheckResponse, error) {
	checkCheck(config)

	expirationsPath := getExpirationsFilePath(config.GlobalConfig)
	if expirationsPath == "" {
		return Untracked, nil
	}

	records, err := readRecordsFromFile(expirationsPath)
	if err != nil {
		return Untracked, err
	}

	rec, ok := records.getFirst(func(rec ExpirationRecord) bool {
		return rec.Target == config.Target
	})

	if ok {
		if rec.Expires.After(time.Now()) {
			return TrackedUnexpired, nil
		} else {
			return TrackedExpired, nil
		}
	} else {
		return Untracked, nil
	}
}
