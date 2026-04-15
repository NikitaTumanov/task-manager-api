CREATE TABLE IF NOT EXISTS instructions (
	id BIGSERIAL PRIMARY KEY,
    instruction TEXT NOT NULL,
    instruction_value INTEGER,
	created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
	updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    task_id BIGINT REFERENCES tasks(id)
);
