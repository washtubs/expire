package expire

import (
	"errors"
	"time"
)

type NewConfig struct {
	GlobalConfig
	BatchRunConfig
	DryRunConfig
	TargetConfig
	Init         bool
	Duration     time.Duration
	ResetOnTouch bool
	NoShadow     bool
}

func checkNew(config *NewConfig) error {
	if config.Target == "" {
		return errors.New("No target")
	}
	return nil
}

func New(config *NewConfig) error {
	err := checkNew(config)
	if err != nil {
		return err
	}

	expirationsPath := getExpirationsFilePath(config.GlobalConfig)

	if expirationsPath == "" {
		if config.Init {
			err := Init(&InitConfig{
				config.GlobalConfig,
				config.DryRunConfig,
			})
			if err != nil {
				return err
			}

			expirationsPath = getExpirationsFilePath(config.GlobalConfig)
		}
	}

	duration := config.Duration
	if duration == 0 {
		duration = DefaultDuration()
	}

	record := &ExpirationRecord{
		Target:       config.Target,
		Expires:      time.Now().Add(duration),
		Duration:     duration,
		ResetOnTouch: config.ResetOnTouch,
	}

	if config.IsDryRun {
		dryRunReporter.ReportAction("Would insert %#v", record)
		return nil
	}

	if expirationsPath == "" {
		return errors.New("No expirations file. Use init or the init config option to create one")
	}

	fp := getExpirationsFilePath(config.GlobalConfig)
	records, err := readRecordsFromFile(fp)
	if err != nil {
		return err
	}

	if config.NoShadow {
		_, exists := records.getFirst(func(rec ExpirationRecord) bool {
			return rec.Target == config.Target
		})
		if exists {
			if config.IsBatchRun {
				return nil
			} else {
				return errors.New("A record already exists and 'no shadow' was requested. Bailing.")
			}
		}
	}
	records.insert(record)

	return writeRecordsToFile(fp, records)
}
