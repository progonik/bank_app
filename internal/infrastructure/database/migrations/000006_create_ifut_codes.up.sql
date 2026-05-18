CREATE TABLE ifut_codes (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

ALTER TABLE entrepreneurs
    DROP COLUMN ifut_code,
    ADD COLUMN ifut_code_id UUID REFERENCES ifut_codes(id);
