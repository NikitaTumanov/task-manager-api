CREATE TABLE IF NOT EXISTS instructions (
	id BIGSERIAL PRIMARY KEY,
    scenario TEXT NOT NULL,
    scenario_value INTEGER,
    next_task_date TIMESTAMPTZ NOT NULL,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    task_id BIGINT UNIQUE REFERENCES tasks(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_instructions_next_task_date ON instructions (next_task_date);