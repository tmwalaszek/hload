package time_formats

import (
	"errors"
	"fmt"
	"time"
)

var timeFormats = []string{
	"02.01.2006",     // DD.MM.YYYY
	"02.01.06",       // DD.MM.YY
	"01/02/2006",     // MM/DD/YYYY
	"01/02/06",       // MM/DD/YY
	"01022006",       // MMDDYYYY
	"010206",         // MMDDYY
	"15:05_20060102", // HH:MM_YYYYMMDD
	"20060102",       // YYYYMMDD
	"01/02/06",       //  MM/DD/YY
}

func timeFromDuration(now time.Time, timeVal string) (int64, error) {
	var d time.Duration
	var err error

	d, err = time.ParseDuration(timeVal)
	if err == nil {
		diff := now.Add(-d)
		return diff.Unix(), nil
	}

	return 0, err
}

func timeZoneOffset() (time.Duration, error) {
	utcTime := time.Now().UTC()
	localTimeZone, err := time.LoadLocation("Europe/Warsaw")
	if err != nil {
		return 0, fmt.Errorf("could not load local timezone: %w", err)
	}

	_, offset := utcTime.In(localTimeZone).Zone()

	return time.Duration(offset) * time.Second, nil
}

func TimeToEpoch(timeVal string) (int64, error) {
	var t time.Time
	var parseErr *time.ParseError
	var err error

	// First we will check if the time is a duration string
	now := time.Now().UTC()
	durationTime, err := timeFromDuration(now, timeVal)
	if err == nil {
		return durationTime, nil
	}

	for _, format := range timeFormats {
		t, err = time.Parse(format, timeVal)
		if err == nil {
			break
		}

		if !errors.As(err, &parseErr) {
			break
		}

		parseErr = nil
	}

	if parseErr != nil {
		return 0, errors.New("invalid time format")
	}

	if err != nil {
		return 0, fmt.Errorf("could not parse time: %w", err)
	}

	timeDiff, err := timeZoneOffset()
	t = t.Add(-timeDiff)
	return t.Unix(), nil
}
