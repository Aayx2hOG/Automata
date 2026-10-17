package scheduler

import (
	"fmt"
	"time"
	_ "time/tzdata"
)

// ScheduleOptions describes a recurring schedule without requiring cron syntax.
type ScheduleOptions struct {
	Frequency string `json:"frequency"`
	Time      string `json:"time"`
	Weekday   int    `json:"weekday"`
	Day       int    `json:"day"`
	Minute    int    `json:"minute"`
	Timezone  string `json:"timezone"`
}

func (o ScheduleOptions) CronExpression() (string, error) {
	if o.Timezone == "" {
		o.Timezone = "UTC"
	}
	if _, err := time.LoadLocation(o.Timezone); err != nil {
		return "", fmt.Errorf("choose a valid time zone")
	}
	prefix := "CRON_TZ=" + o.Timezone + " "
	if o.Frequency == "hourly" {
		if o.Minute < 0 || o.Minute > 59 {
			return "", fmt.Errorf("minute must be between 0 and 59")
		}
		return prefix + fmt.Sprintf("%d * * * *", o.Minute), nil
	}
	t, err := time.Parse("15:04", o.Time)
	if err != nil {
		return "", fmt.Errorf("choose a valid time in HH:MM format")
	}
	switch o.Frequency {
	case "daily":
		return prefix + fmt.Sprintf("%d %d * * *", t.Minute(), t.Hour()), nil
	case "weekly":
		if o.Weekday < 0 || o.Weekday > 6 {
			return "", fmt.Errorf("choose a valid weekday")
		}
		return prefix + fmt.Sprintf("%d %d * * %d", t.Minute(), t.Hour(), o.Weekday), nil
	case "monthly":
		if o.Day < 1 || o.Day > 28 {
			return "", fmt.Errorf("day must be between 1 and 28 so the schedule runs every month")
		}
		return prefix + fmt.Sprintf("%d %d %d * *", t.Minute(), t.Hour(), o.Day), nil
	default:
		return "", fmt.Errorf("choose hourly, daily, weekly or monthly")
	}
}
