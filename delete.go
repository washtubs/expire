package expire

import "errors"
import "os"

type DeleteConfig struct {
	GlobalConfig
	BatchRunConfig
	DryRunConfig
	TargetConfig
	DeInit bool
}

func checkDelete(config *DeleteConfig) error {
	if config.Target == "" {
		return errors.New("No target")
	}
	return nil
}

func Delete(config *DeleteConfig) error {
	checkDelete(config)

	expirationsPath := getExpirationsFilePath(config.GlobalConfig)
	if expirationsPath == "" {
		if config.IsBatchRun {
			return nil
		} else {
			return errors.New("No expirations file")
		}
	}

	records, err := readRecordsFromFile(expirationsPath)
	if err != nil {
		return err
	}

	_, present := records.deleteFirst(func(rec ExpirationRecord) bool {
		return rec.Target == config.Target
	})

	if !present {
		if config.IsDryRun {
			dryRunReporter.ReportAction("Will not delete non-existent record: %s", config.Target)
		}
		if config.IsBatchRun {
			return nil
		} else {
			return errors.New("No such record: " + config.Target)
		}
	}

	if config.IsDryRun {
		dryRunReporter.ReportAction("Will delete record: %s", config.Target)
		if len(records) == 0 && config.DeInit {
			dryRunReporter.ReportAction("Will delete the file: %s", expirationsPath)
		}
		return nil
	}

	if len(records) == 0 && config.DeInit {
		return os.Remove(expirationsPath)
	} else {
		return writeRecordsToFile(expirationsPath, records)
	}
}
