package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	instructiondomain "example.com/taskservice/internal/domain/instruction"
)

type InstructionRepository struct {
	pool *pgxpool.Pool
}

func NewInstructionRepository(pool *pgxpool.Pool) *InstructionRepository {
	return &InstructionRepository{pool: pool}
}

func (r *InstructionRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

func (r *InstructionRepository) Create(ctx context.Context, tx pgx.Tx, instruction *instructiondomain.Instruction) (*instructiondomain.Instruction, error) {
	const query = `
		INSERT INTO instructions (scenario, scenario_value, next_task_date, created_at, updated_at, task_id)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, scenario, scenario_value, next_task_date, created_at, updated_at, task_id
	`

	row := tx.QueryRow(ctx, query, instruction.Scenario, instruction.ScenarioValue, instruction.NextTaskDate, instruction.CreatedAt, instruction.UpdatedAt, instruction.TaskID)
	created, err := scanInstruction(row)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *InstructionRepository) GetByID(ctx context.Context, id int64) (*instructiondomain.Instruction, error) {
	const query = `
		SELECT id, scenario, scenario_value, next_task_date, created_at, updated_at, task_id
		FROM instructions
		WHERE id = $1
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

func (r *InstructionRepository) ListByNextTaskDate(ctx context.Context, date time.Time) ([]instructiondomain.Instruction, error) {
	startDate := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 0, 1)

	const query = `
		SELECT id, scenario, scenario_value, next_task_date, created_at, updated_at, task_id
		FROM instructions
		WHERE next_task_date >= $1 AND next_task_date < $2
	`

	rows, err := r.pool.Query(ctx, query, startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	instructions := make([]instructiondomain.Instruction, 0)
	for rows.Next() {
		instruction, err := scanInstruction(rows)
		if err != nil {
			return nil, err
		}

		instructions = append(instructions, *instruction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return instructions, nil
}

func (r *InstructionRepository) Update(ctx context.Context, tx pgx.Tx, instruction *instructiondomain.Instruction) (*instructiondomain.Instruction, error) {
	var row pgx.Row

	const query = `
		UPDATE instructions
		SET scenario = $1,
		    scenario_value = $2,
		    next_task_date = $3,
			updated_at = $4,
			task_id = $5
		WHERE id = $6
		RETURNING id, scenario, scenario_value, next_task_date, created_at, updated_at, task_id
	`
	row = tx.QueryRow(ctx, query, instruction.Scenario, instruction.ScenarioValue, instruction.NextTaskDate, instruction.UpdatedAt, instruction.TaskID, instruction.ID)

	updated, err := scanInstruction(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, instructiondomain.ErrNotFound
		}

		return nil, err
	}

	return updated, nil
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

func (r *InstructionRepository) List(ctx context.Context) ([]instructiondomain.Instruction, error) {
	const query = `
		SELECT id, scenario, scenario_value, next_task_date, created_at, updated_at, task_id
		FROM instructions
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	instructions := make([]instructiondomain.Instruction, 0)
	for rows.Next() {
		instruction, err := scanInstruction(rows)
		if err != nil {
			return nil, err
		}

		instructions = append(instructions, *instruction)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return instructions, nil
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
		&scenario,
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
