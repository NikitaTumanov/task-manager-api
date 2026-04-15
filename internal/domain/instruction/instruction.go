package instruction

import "time"

type Scenario string

const (
	ScenarioZero          Scenario = ""
	ScenarioDaily         Scenario = "daily"
	ScenarioMonthly       Scenario = "monthly"
	ScenarioSpecificDates Scenario = "specific_dates"
	ScenarioEven          Scenario = "even"
	ScenarioOdd           Scenario = "odd"
)

type Instruction struct {
	ID            int64     `json:"id"`
	Scenario      Scenario  `json:"scenario"`
	ScenarioValue int       `json:"scenario_value"`
	NextTaskDate  time.Time `json:"next_task_date"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	TaskID        int64     `json:"task_id"`
}

func (s Scenario) Valid() bool {
	switch s {
	case ScenarioZero, ScenarioDaily, ScenarioMonthly, ScenarioSpecificDates, ScenarioEven, ScenarioOdd:
		return true
	default:
		return false
	}
}
