package postgres

import (
	"context"
	"errors"

	instructiondomain "example.com/taskservice/internal/domain/instruction"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type InstructionRepository struct {
	pool *pgxpool.Pool
}

func NewInstructionRepository(pool *pgxpool.Pool) *InstructionRepository {
	return &InstructionRepository{pool: pool}
}

func (r *InstructionRepository) Create(ctx context.Context, instruction *instructiondomain.Instruction) (*instructiondomain.Instruction, error) {
	const query = `
		INSERT INTO instructions (scenario, scenario_value, next_task_date, created_at, updated_at, task_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, scenario, scenario_value, next_task_date, created_at, updated_at, task_id
	`

	row := r.pool.QueryRow(ctx, query, instruction.Scenario, instruction.ScenarioValue, instruction.NextTaskDate, instruction.CreatedAt, instruction.UpdatedAt, instruction.TaskID)
	created, err := scanInstruction(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *InstructionRepository) GetByTaskID(ctx context.Context, id int64) (*instructiondomain.Instruction, error) {
	const query = `
		SELECT id, scenario, scenario_value, next_task_date, created_at, updated_at, task_id
		FROM instructions
		WHERE task_id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)
	found, err := scanInstruction(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, instructiondomain.ErrNotFound
		}

		return nil, err
	}

	return found, nil
}

func (r *InstructionRepository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM instructions WHERE id = $1`

	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return instructiondomain.ErrNotFound
	}

	return nil
}

type instructionScanner interface {
	Scan(dest ...any) error
}

func scanInstruction(scanner instructionScanner) (*instructiondomain.Instruction, error) {
	var (
		instruction instructiondomain.Instruction
		scenario    string
	)

	if err := scanner.Scan(
		&instruction.ID,
		&instruction.ScenarioValue,
		&instruction.NextTaskDate,
		&instruction.CreatedAt,
		&instruction.UpdatedAt,
		&instruction.TaskID,
	); err != nil {
		return nil, err
	}

	instruction.Scenario = instructiondomain.Scenario(scenario)

	return &instruction, nil
}
