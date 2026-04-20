package task

import (
	"context"
	"fmt"
	"strings"
	"time"

	instructiondomain "example.com/taskservice/internal/domain/instruction"
	taskdomain "example.com/taskservice/internal/domain/task"
	"example.com/taskservice/internal/scheduler"
)

var (
	titleLength       = 150
	descriptionLength = 2000
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

	if normalized.Scenario != instructiondomain.ScenarioZero && normalized.Scenario != instructiondomain.ScenarioSpecificDates {
		nextTaskDate := scheduler.CalculateNextDate(normalized.Deadline, normalized.Scenario, normalized.ScenarioValue)

		model.Deadline = nextTaskDate
		model.Status = taskdomain.StatusNew

		nextTask, err := s.repo.Create(ctx, model)
		if err != nil {
			return nil, err
		}

		instructionModel := &instructiondomain.Instruction{
			Scenario:      normalized.Scenario,
			ScenarioValue: normalized.ScenarioValue,
			NextTaskDate:  nextTaskDate,
			CreatedAt:     now,
			UpdatedAt:     now,
			TaskID:        nextTask.ID,
		}

		_, err = s.instructionRepo.Create(ctx, instructionModel)
		if err != nil {
			return nil, err
		}
	} else if normalized.Scenario == instructiondomain.ScenarioSpecificDates {
		for _, d := range input.SpecificDates {
			model.Deadline = d
			model.Status = taskdomain.StatusNew
			_, err := s.repo.Create(ctx, model)
			if err != nil {
				return nil, err
			}
		}
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

	if len([]rune(input.Title)) > titleLength {
		return CreateInput{}, fmt.Errorf("%w: title is too long", ErrInvalidInput)
	}

	if len([]rune(input.Description)) > descriptionLength {
		return CreateInput{}, fmt.Errorf("%w: description is too long", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if input.Deadline.IsZero() {
		return CreateInput{}, fmt.Errorf("%w: deadline is required", ErrInvalidInput)
	}
	if !input.Deadline.After(time.Now().UTC()) {
		return CreateInput{}, fmt.Errorf("%w: deadline must refer to the future", ErrInvalidInput)
	}

	if !input.Scenario.Valid() {
		return CreateInput{}, fmt.Errorf("%w: invalid scenario", ErrInvalidInput)
	}

	switch input.Scenario {
	case instructiondomain.ScenarioDaily:
		if input.ScenarioValue <= 0 {
			return CreateInput{}, fmt.Errorf("%w: scenario value is required (value > 0)", ErrInvalidInput)
		}

	case instructiondomain.ScenarioMonthly:
		if input.ScenarioValue <= 0 || input.ScenarioValue > 31 {
			return CreateInput{}, fmt.Errorf("%w: scenario value is required (0 < value <= 31)", ErrInvalidInput)
		}

	case instructiondomain.ScenarioSpecificDates:
		if len(input.SpecificDates) == 0 {
			return CreateInput{}, fmt.Errorf("%w: specific dates is required", ErrInvalidInput)
		}
	}

	return input, nil
}

func validateUpdateInput(input UpdateInput) (UpdateInput, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Description = strings.TrimSpace(input.Description)

	if input.Title == "" {
		return UpdateInput{}, fmt.Errorf("%w: title is required", ErrInvalidInput)
	}

	if len([]rune(input.Title)) > titleLength {
		return UpdateInput{}, fmt.Errorf("%w: title is too long", ErrInvalidInput)
	}

	if len([]rune(input.Description)) > descriptionLength {
		return UpdateInput{}, fmt.Errorf("%w: description is too long", ErrInvalidInput)
	}

	if !input.Status.Valid() {
		return UpdateInput{}, fmt.Errorf("%w: invalid status", ErrInvalidInput)
	}

	if !input.Deadline.After(time.Now().UTC()) && !input.Deadline.IsZero() {
		return UpdateInput{}, fmt.Errorf("%w: deadline must refer to the future", ErrInvalidInput)
	}

	return input, nil
}
