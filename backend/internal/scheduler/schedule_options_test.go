package scheduler

import (
	"go.uber.org/zap"
	"testing"
	"time"
)

func TestScheduleOptionsNextRun(t *testing.T) {
	after := time.Date(2026, 10, 17, 0, 0, 0, 0, time.UTC)
	tests := []struct {
		name    string
		options ScheduleOptions
		want    string
	}{
		{"daily local time", ScheduleOptions{Frequency: "daily", Time: "09:30", Timezone: "Asia/Kolkata"}, "2026-10-17T04:00:00Z"},
		{"weekly Monday", ScheduleOptions{Frequency: "weekly", Time: "10:15", Weekday: 1, Timezone: "UTC"}, "2026-10-19T10:15:00Z"},
		{"monthly", ScheduleOptions{Frequency: "monthly", Time: "08:00", Day: 1}, "2026-11-01T08:00:00Z"},
		{"hourly", ScheduleOptions{Frequency: "hourly", Minute: 45}, "2026-10-17T00:45:00Z"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expression, err := tt.options.CronExpression()
			if err != nil {
				t.Fatal(err)
			}
			next, err := New(nil, zap.NewNop()).ParseNext(expression, after)
			if err != nil {
				t.Fatal(err)
			}
			if got := next.UTC().Format(time.RFC3339); got != tt.want {
				t.Fatalf("got %s, want %s", got, tt.want)
			}
		})
	}
}
func TestScheduleOptionsRejectInvalidInput(t *testing.T) {
	for _, options := range []ScheduleOptions{
		{Frequency: "daily", Time: "25:00"}, {Frequency: "daily", Time: "09:00", Timezone: "invalid/zone"},
		{Frequency: "weekly", Time: "09:00", Weekday: 7}, {Frequency: "monthly", Time: "09:00", Day: 31},
		{Frequency: "hourly", Minute: 60}, {Frequency: "invalid", Time: "09:00"},
	} {
		if _, err := options.CronExpression(); err == nil {
			t.Errorf("accepted invalid options: %+v", options)
		}
	}
}
