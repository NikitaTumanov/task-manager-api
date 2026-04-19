package instruction

import (
	"context"
	"time"

	instructiondomain "example.com/taskservice/internal/domain/instruction"
	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository interface {
	Create(ctx context.Context, instruction *instructiondomain.Instruction) (*instructiondomain.Instruction, error)
	GetByID(ctx context.Context, id int64) (*instructiondomain.Instruction, error)
	GetByTaskID(ctx context.Context, id int64) (*instructiondomain.Instruction, error)
	Update(ctx context.Context, instruction *instructiondomain.Instruction) (*instructiondomain.Instruction, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]instructiondomain.Instruction, error)
}

type TaskRepository interface {
	Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
}

type UseCase interface {
	Create(ctx context.Context, input CreateInput) (*instructiondomain.Instruction, error)
	GetByID(ctx context.Context, id int64) (*instructiondomain.Instruction, error)
	GetByTaskID(ctx context.Context, id int64) (*instructiondomain.Instruction, error)
	Update(ctx context.Context, id int64, input UpdateInput) (*instructiondomain.Instruction, error)
	Delete(ctx context.Context, id int64) error
	List(ctx context.Context) ([]instructiondomain.Instruction, error)
}

type CreateInput struct {
	Scenario      instructiondomain.Scenario
	ScenarioValue int
	TaskID        int64
	SpecificDates []time.Time
}

type UpdateInput struct {
	Scenario      instructiondomain.Scenario
	ScenarioValue int
}
