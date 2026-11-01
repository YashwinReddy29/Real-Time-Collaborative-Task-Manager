-- Migration 002: Enforce task/column ownership consistency.
-- Safe to reapply. Existing mismatched records must be resolved
-- before this migration can complete.

CREATE UNIQUE INDEX IF NOT EXISTS
    idx_columns_id_board_unique
ON columns (id, board_id);

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1
        FROM pg_constraint
        WHERE conname = 'tasks_column_board_fkey'
          AND conrelid = 'tasks'::regclass
    ) THEN
        ALTER TABLE tasks
            ADD CONSTRAINT tasks_column_board_fkey
            FOREIGN KEY (column_id, board_id)
            REFERENCES columns (id, board_id)
            ON DELETE CASCADE;
    END IF;
END
$$;
