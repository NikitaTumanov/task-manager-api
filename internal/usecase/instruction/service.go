package instruction

import (
	"context"
	"fmt"
	"time"

	"example.com/taskservice/internal/domain/instruction"
	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/scheduler"
)

type Service struct {
	repo     Repository
	taskRepo TaskRepository
	now      func() time.Time
}

func NewService(repo Repository, taskRepo TaskRepository) *Service {
	return &Service{
		repo:     repo,
		taskRepo: taskRepo,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*instruction.Instruction, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	task, err := s.taskRepo.GetByID(ctx, normalized.TaskID)
	if err != nil {
		return nil, err
	}

	//TODO отдельно обработку для ScenarioSpecificDates
	var created *instruction.Instruction
	if normalized.Scenario != instruction.ScenarioZero && normalized.Scenario != instruction.ScenarioSpecificDates {
		_, err = s.repo.GetByTaskID(ctx, normalized.TaskID)
		if err == nil {
			return nil, fmt.Errorf("%w: instruction for task id %d already exists", ErrInvalidInput, normalized.TaskID)
		}

		nextTaskDate := scheduler.CalculateNextDate(task.Deadline, input.Scenario, input.ScenarioValue)

		task.Deadline = nextTaskDate
		task.Status = taskdomain.StatusNew
		newTask, _ := s.taskRepo.Create(ctx, task)

		model := &instruction.Instruction{
			Scenario:      normalized.Scenario,
			ScenarioValue: normalized.ScenarioValue,
			NextTaskDate:  nextTaskDate,
			TaskID:        newTask.ID,
		}
		now := s.now()
		model.CreatedAt = now
		model.UpdatedAt = now

		created, err = s.repo.Create(ctx, model)
		if err != nil {
			return nil, err
		}
	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*instruction.Instruction, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByTaskID(ctx context.Context, id int64) (*instruction.Instruction, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: task id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByTaskID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*instruction.Instruction, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	model, err := s.repo.GetByTaskID(ctx, id)
	if err != nil {
		return nil, err
	}

	nextTaskDate := scheduler.CalculateNextDate(task.Deadline, input.Scenario, input.ScenarioValue)

	task.Deadline = nextTaskDate
	task.Status = taskdomain.StatusNew
	newTask, _ := s.taskRepo.Create(ctx, task)

	model.Scenario = normalized.Scenario
	model.ScenarioValue = normalized.ScenarioValue
	model.NextTaskDate = nextTaskDate
	model.TaskID = newTask.ID

	now := s.now()
	model.UpdatedAt = now

	updated, err := s.repo.Update(ctx, model)
	if err != nil {
		return nil, err
	}

	return updated, nil
}

func (s *Service) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.Delete(ctx, id)
}

func (s *Service) List(ctx context.Context) ([]instruction.Instruction, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	if !input.Scenario.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid scenario", ErrInvalidInput)
	}

	switch input.Scenario {
	case instruction.ScenarioDaily:
		if input.ScenarioValue <= 0 {
			return CreateInput{}, fmt.Errorf("%w: scenario value is required (value > 0)", ErrInvalidInput)
		}

	case instruction.ScenarioMonthly:
		if input.ScenarioValue <= 0 || input.ScenarioValue > 31 {
			return CreateInput{}, fmt.Errorf("%w: scenario value is required (0 < value <= 31)", ErrInvalidInput)
		}

	case instruction.ScenarioSpecificDates:
		if len(input.SpecificDates) == 0 {
			return CreateInput{}, fmt.Errorf("%w: specific dates is required", ErrInvalidInput)
		}
	}

	if input.TaskID <= 0 {
		return CreateInput{}, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	now := time.Now().UTC()
	if len(input.SpecificDates) > 0 {
		for _, d := range input.SpecificDates {
			if !d.After(now) {
				return CreateInput{}, fmt.Errorf("%w: specific dates must refer to the future", ErrInvalidInput)
			}
		}
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	if !input.Scenario.Valid() || input.Scenario == instruction.ScenarioSpecificDates {
		return UpdateInput{}, fmt.Errorf("%w: invalid scenario", ErrInvalidInput)
	}

	switch input.Scenario {
	case instruction.ScenarioDaily:
		if input.ScenarioValue <= 0 {
			return UpdateInput{}, fmt.Errorf("%w: scenario value is required (value > 0)", ErrInvalidInput)
		}

	case instruction.ScenarioMonthly:
		if input.ScenarioValue <= 0 || input.ScenarioValue > 31 {
			return UpdateInput{}, fmt.Errorf("%w: scenario value is required (0 < value <= 31)", ErrInvalidInput)
		}

	}

	return input, nil
}
