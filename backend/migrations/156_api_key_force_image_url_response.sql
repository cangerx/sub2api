-- Add per-API-key control for forcing OpenAI image APIs to return URL responses.
ALTER TABLE api_keys
    ADD COLUMN IF NOT EXISTS force_image_url_response BOOLEAN NOT NULL DEFAULT FALSE;
