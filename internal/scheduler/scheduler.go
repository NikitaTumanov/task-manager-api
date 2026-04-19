package scheduler

import (
	"time"

	"example.com/taskservice/internal/domain/instruction"
)

func CalculateNextDate(date time.Time, scenario instruction.Scenario, value int) time.Time {
	switch scenario {
	case instruction.ScenarioDaily:
		return date.AddDate(0, 0, value)

	case instruction.ScenarioMonthly:
		if date.Day() < value {
			return normalizeDate(date.Year(), date.Month(), value, date)
		}
		return normalizeDate(date.Year(), date.Month()+1, value, date)

	case instruction.ScenarioEven:
		nextDay := date.AddDate(0, 0, 1)
		for nextDay.Day()%2 != 0 {
			nextDay = nextDay.AddDate(0, 0, 1)
		}
		return nextDay

	case instruction.ScenarioOdd:
		nextDay := date.AddDate(0, 0, 1)
		for nextDay.Day()%2 == 0 {
			nextDay = nextDay.AddDate(0, 0, 1)
		}
		return nextDay
	}

	return time.Time{}
}

func normalizeDate(year int, month time.Month, day int, t time.Time) time.Time {
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, t.Location()).Day()

	if day > lastDay {
		day = lastDay
	}

	return time.Date(year, month, day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), t.Location())
}
