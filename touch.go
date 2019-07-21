package expire

import (
	"time"
)

type TouchConfig struct {
	GlobalConfig
	BatchRunConfig
	DryRunConfig
	TargetConfig
}

func Touch(config *TouchConfig) error {
	return Update(&UpdateConfig{
		config.GlobalConfig,
		config.BatchRunConfig,
		config.DryRunConfig,
		config.TargetConfig,
	}, func(rec *ExpirationRecord) {
		// touch this record: i.e. if it has not expired, reset the timer
		if rec.ResetOnTouch && rec.Expires.After(time.Now()) {
			rec.Expires = time.Now().Add(rec.Duration)
		}
	})

}
