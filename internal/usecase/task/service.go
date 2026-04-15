package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	"example.com/taskservice/internal/domain/instruction"
	instructiondomain "example.com/taskservice/internal/domain/instruction"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type Service struct {
	repo            Repository
	instructionRepo InstructionRepository
	now             func() time.Time
}

func NewService(repo Repository, instructionRepo InstructionRepository) *Service {
	return &Service{
		repo:            repo,
		instructionRepo: instructionRepo,
		now:             func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*taskdomain.Task, error) {
	normalized, err := validateCreateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Deadline:    normalized.Deadline,
	}
	now := s.now()
	model.CreatedAt = now
	model.UpdatedAt = now

	created, err := s.repo.Create(ctx, model)
	if err != nil {
		return nil, err
	}

	//TODO отдельно обработку для ScenarioSpecificDates
	if normalized.Scenario != instruction.ScenarioZero && normalized.Scenario != instruction.ScenarioSpecificDates {
		nextTaskDate := calculateNextDate(normalized.Deadline, normalized.Scenario, normalized.ScenarioValue)

		model.Deadline = nextTaskDate
		//TODO добавить обработку не созданной задачи
		nextTask, _ := s.repo.Create(ctx, model)

		instructionModel := &instructiondomain.Instruction{
			Scenario:      normalized.Scenario,
			ScenarioValue: normalized.ScenarioValue,
			NextTaskDate:  nextTaskDate,
			CreatedAt:     now,
			UpdatedAt:     now,
			TaskID:        nextTask.ID,
		}

		//TODO добавить обработку не созданной инструкции
		_, err = s.instructionRepo.Create(ctx, instructionModel)

	}

	return created, nil
}

func (s *Service) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id int64, input UpdateInput) (*taskdomain.Task, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: id must be positive", ErrInvalidInput)
	}

	normalized, err := validateUpdateInput(input)
	if err != nil {
		return nil, err
	}

	model := &taskdomain.Task{
		ID:          id,
		Title:       normalized.Title,
		Description: normalized.Description,
		Status:      normalized.Status,
		Deadline:    normalized.Deadline,
		UpdatedAt:   s.now(),
	}

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

func (s *Service) List(ctx context.Context) ([]taskdomain.Task, error) {
	return s.repo.List(ctx)
}

func validateCreateInput(input CreateInput) (CreateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return CreateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if input.Status == "" {
		input.Status = taskdomain.StatusNew
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if input.Deadline.IsZero() {
		return CreateInput{}, fmt.Errorf("%w: deadline is required", ErrInvalidInput)
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

	if !input.Scenario.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid scenario", ErrInvalidInput)
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	return input, nil
}

func calculateNextDate(date time.Time, scenario instruction.Scenario, value int) time.Time {
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
