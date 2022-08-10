package time_utils

import (
	"fmt"
	"time"
)

func FormatInZone(t *time.Time, zone int, format string) (timeStr string) {
	loc := time.FixedZone("", zone*3600)
	zoneTime := t.In(loc)
	timeStr = zoneTime.Format(format)
	return timeStr
}

func TruncateHour(timezone string, t time.Time) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return t, err
	}
	return t.In(loc).Truncate(time.Hour), nil
}

func GetDayHourStr(t *time.Time, timezone string) (string, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return "", err
	}
	newT := t.In(loc)
	return fmt.Sprintf("%d-%d-%dT%.2d", newT.Year(), newT.Month(), newT.Day(), newT.Hour()), nil
}

func GetLastHourStr(t *time.Time, timezone string) (string, error) {
	lastHour := t.Add(-time.Hour)
	return GetDayHourStr(&lastHour, timezone)
}

func GetFirstDateOfWeek(timezone string) string {
	now := time.Now()
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return ""
	}
	newT := now.In(loc)
	offset := int(time.Sunday - newT.Weekday())
	weekStartDate := time.Date(newT.Year(), newT.Month(), newT.Day(), 0, 0, 0, 0, time.Local).AddDate(0, 0, offset)
	return weekStartDate.Format("2006-01-02")
}
