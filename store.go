package expire

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"time"
)

const dateTimeFormat string = time.RFC3339

func readRecords(reader io.Reader) (ExpirationRecords, error) {
	r := csv.NewReader(reader)
	records, err := r.ReadAll()
	if err != nil {
		// IO / CSV parse error?
		return nil, err
	}
	out := make(ExpirationRecords, 0, len(records))
	for i, r := range records {
		if i == 0 {
			// Skip the header
			continue
		}
		rec, err := fromRecord(r)
		if err != nil {
			// malformed record?
			return nil, errors.New(fmt.Sprintf("Error parsing expirations file (line #%d): %s", i, err))
		}

		out = append(out, rec)
	}
	return out, nil
}

func writeRecords(writer io.Writer, recs ExpirationRecords) error {
	w := csv.NewWriter(writer)

	w.Write([]string{"target", "expires", "duration", "resetOnTouch"})

	for _, r := range recs {
		err := w.Write(toRecord(*r))
		if err != nil {
			return err
		}
	}

	w.Flush()

	return nil
}

func readRecordsFromFile(expirationsFile string) (ExpirationRecords, error) {
	f, err := os.Open(expirationsFile)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	recs, err := readRecords(f)
	if err != nil {
		return nil, err
	}
	return recs, f.Sync()
}

func writeRecordsToFile(expirationsFile string, records ExpirationRecords) error {
	f, err := os.Create(expirationsFile)
	if err != nil {
		return err
	}
	err = writeRecords(f, records)
	if err != nil {
		return err
	}
	return f.Sync()
}

func fromRecord(r []string) (*ExpirationRecord, error) {

	if len(r) != 4 {
		return nil, errors.New("Incorrect length record. Should be 4")
	}

	expires, err := time.Parse(dateTimeFormat, r[1])
	if err != nil {
		return nil, err
	}

	duration, err := ParseDurationString(r[2])
	if err != nil {
		return nil, err
	}

	resetOnTouch := false
	if r[3] == "yes" {
		resetOnTouch = true
	}

	return &ExpirationRecord{
		Target:       r[0],
		Expires:      expires,
		Duration:     duration,
		ResetOnTouch: resetOnTouch,
	}, nil
}

func toRecord(e ExpirationRecord) []string {
	reset := "yes"
	if !e.ResetOnTouch {
		reset = "no"
	}
	expires := e.Expires.Format(dateTimeFormat)

	return []string{
		e.Target,
		string(expires),
		e.Duration.String(),
		reset,
	}
}
