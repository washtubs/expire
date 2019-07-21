package expire

import (
	"os"
)

type InitConfig struct {
	GlobalConfig
	DryRunConfig
}

// Initializes an expirations file in the current directory
// If the file already exists, it will do nothing
func Init(config *InitConfig) error {
	_, err := os.Stat(config.getFileName())
	if !os.IsNotExist(err) {
		if config.IsDryRun {
			dryRunReporter.ReportAction("File exists: %s. Will not re-initialize.", config.getFileName())
		}
		return nil
	}

	if config.IsDryRun {
		dryRunReporter.ReportAction("Would create %s", config.getFileName())
		return nil
	}

	return writeRecordsToFile(config.getFileName(), ExpirationRecords{})
}
