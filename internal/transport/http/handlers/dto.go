package handlers

import (
	"time"

	"example.com/taskservice/internal/domain/instruction"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type taskMutationDTO struct {
	Title         string               `json:"title"`
	Description   string               `json:"description"`
	Status        taskdomain.Status    `json:"status"`
	Deadline      time.Time            `json:"deadline"`
	Scenario      instruction.Scenario `json:"scenario"`
	ScenarioValue int                  `json:"scenario_value"`
	SpecificDates []time.Time          `json:"specific_dates"`
}

type instructionMutationDTO struct {
	Scenario      instruction.Scenario `json:"scenario"`
	ScenarioValue int                  `json:"scenario_value"`
	SpecificDates []time.Time          `json:"specific_dates"`
	TaskID        int64                `json:"task_id"`
}

type taskDTO struct {
	ID          int64             `json:"id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Status      taskdomain.Status `json:"status"`
	Deadline    time.Time         `json:"deadline"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type instructionDTO struct {
	ID            int64                `json:"id"`
	Scenario      instruction.Scenario `json:"scenario"`
	ScenarioValue int                  `json:"scenario_value"`
	NextTaskDate  time.Time            `json:"next_task_date"`
	CreatedAt     time.Time            `json:"created_at"`
	UpdatedAt     time.Time            `json:"updated_at"`
	TaskID        int64                `json:"task_id"`
}

func newTaskDTO(task *taskdomain.Task) taskDTO {
	return taskDTO{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
		Deadline:    task.Deadline,
		CreatedAt:   task.CreatedAt,
		UpdatedAt:   task.UpdatedAt,
	}
}

func newInstructionDTO(instruction *instruction.Instruction) instructionDTO {
	return instructionDTO{
		ID:            instruction.ID,
		Scenario:      instruction.Scenario,
		ScenarioValue: instruction.ScenarioValue,
		NextTaskDate:  instruction.NextTaskDate,
		CreatedAt:     instruction.CreatedAt,
		UpdatedAt:     instruction.UpdatedAt,
		TaskID:        instruction.TaskID,
	}
}