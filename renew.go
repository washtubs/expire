package expire

import (
	"time"
)

type RenewConfig struct {
	GlobalConfig
	BatchRunConfig
	DryRunConfig
	TargetConfig
}

func Renew(config *RenewConfig) error {
	return Update(&UpdateConfig{
		config.GlobalConfig,
		config.BatchRunConfig,
		config.DryRunConfig,
		config.TargetConfig,
	}, func(rec *ExpirationRecord) {
		// renew this record: remake it with the same settings
		// i.e. simply reset the timer
		rec.Expires = time.Now().Add(rec.Duration)
	})
}
