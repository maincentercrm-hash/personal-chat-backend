-- Migration: Add unique constraint on message_reads (message_id, user_id)
-- Required for ON CONFLICT clause in CreateRead function

-- Add unique constraint if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'message_reads_message_id_user_id_unique'
    ) THEN
        ALTER TABLE message_reads
        ADD CONSTRAINT message_reads_message_id_user_id_unique
        UNIQUE (message_id, user_id);
    END IF;
END $$;

-- Also create index for better query performance (if not exists)
CREATE INDEX IF NOT EXISTS idx_message_reads_message_user
ON message_reads (message_id, user_id);
