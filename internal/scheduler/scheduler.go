package customscheduler

import (
	"context"
	"log/slog"
	"os"
	"time"

	instructiondomain "example.com/taskservice/internal/domain/instruction"
	taskdomain "example.com/taskservice/internal/domain/task"
	"github.com/jackc/pgx/v5"
)

type TaskRepository interface {
	Create(ctx context.Context, tx pgx.Tx, task *taskdomain.Task) (*taskdomain.Task, error)
	GetByID(ctx context.Context, id int64) (*taskdomain.Task, error)
}

type InstructionRepository interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)
	ListByNextTaskDate(ctx context.Context, date time.Time) ([]instructiondomain.Instruction, error)
	Update(ctx context.Context, tx pgx.Tx, instruction *instructiondomain.Instruction) (*instructiondomain.Instruction, error)
}

type Scheduler struct {
	taskRepo        TaskRepository
	instructionRepo InstructionRepository
	now             func() time.Time
}

func NewScheduler(taskRepo TaskRepository, instructionRepo InstructionRepository) *Scheduler {
	return &Scheduler{
		taskRepo:        taskRepo,
		instructionRepo: instructionRepo,
		now:             time.Now,
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	logger.Info("scheduler started")
	instructions, err := s.instructionRepo.ListByNextTaskDate(ctx, s.now())
	if err != nil {
		return err
	}

	var lastErr error
	for _, instruction := range instructions {
		if err := s.processInstruction(ctx, instruction); err != nil {
			lastErr = err
			continue
		}
	}

	return lastErr
}

func (s *Scheduler) processInstruction(ctx context.Context, instruction instructiondomain.Instruction) error {
	tx, err := s.instructionRepo.BeginTx(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	task, err := s.taskRepo.GetByID(ctx, instruction.TaskID)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	nextTaskDate := CalculateNextDate(task.Deadline, instruction.Scenario, instruction.ScenarioValue)

	now := s.now()
	task.Deadline = nextTaskDate
	task.Status = taskdomain.StatusNew
	task.CreatedAt = now
	task.UpdatedAt = now

	newTask, err := s.taskRepo.Create(ctx, tx, task)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	instruction.NextTaskDate = newTask.Deadline
	instruction.TaskID = newTask.ID
	instruction.UpdatedAt = now

	_, err = s.instructionRepo.Update(ctx, tx, &instruction)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	return tx.Commit(ctx)

}
