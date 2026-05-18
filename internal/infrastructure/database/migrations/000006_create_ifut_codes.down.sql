ALTER TABLE entrepreneurs
    DROP COLUMN ifut_code_id,
    ADD COLUMN ifut_code INT NOT NULL DEFAULT 0;

DROP TABLE IF EXISTS ifut_codes;
