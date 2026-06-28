ALTER TABLE usage_logs
    ADD COLUMN IF NOT EXISTS image_prompt TEXT,
    ADD COLUMN IF NOT EXISTS image_urls JSONB,
    ADD COLUMN IF NOT EXISTS image_revised_prompts JSONB;
