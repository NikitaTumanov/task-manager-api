CREATE TABLE IF NOT EXISTS instructions (
	id BIGSERIAL PRIMARY KEY,
    instruction TEXT NOT NULL,
    instruction_value INTEGER,
    next_task_date TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    task_id BIGINT REFERENCES tasks(id)
);

CREATE INDEX IF NOT EXISTS idx_instructions_next_task_date ON instructions (next_task_date);