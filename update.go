package expire

import "errors"

type UpdateConfig struct {
	GlobalConfig
	BatchRunConfig
	DryRunConfig
	TargetConfig
}

func Update(config *UpdateConfig, action func(*ExpirationRecord)) error {
	if config.Target == "" {
		return errors.New("No target")
	}

	expirationsPath := getExpirationsFilePath(config.GlobalConfig)
	if expirationsPath == "" {
		if config.IsBatchRun {
			// TODO consider maybe throwing an error anyway in the case of a global type error
			return nil
		} else {
			return errors.New("No expirations file")
		}
	}

	records, err := readRecordsFromFile(expirationsPath)
	if err != nil {
		return err
	}

	ok := records.updateFirst(func(rec ExpirationRecord) bool {
		return rec.Target == config.Target
	}, action)

	if !ok {
		if config.IsDryRun {
			dryRunReporter.ReportAction("Will not touch non-existent record: %s", config.Target)
		}
		if config.IsBatchRun {
			return nil
		} else {
			return errors.New("No such record: " + config.Target)
		}
	}

	if config.IsDryRun {
		dryRunReporter.ReportAction("Will touch record: %s", config.Target)
		return nil
	}

	return writeRecordsToFile(expirationsPath, records)
}
