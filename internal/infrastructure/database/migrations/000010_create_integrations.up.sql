CREATE TABLE integrations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    code VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    active BOOLEAN NOT NULL DEFAULT true,
    active_until TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE integration_users (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    integration_id UUID NOT NULL REFERENCES integrations(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(50) NOT NULL DEFAULT 'user',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (integration_id, user_id)
);

INSERT INTO integrations (code, name, active)
VALUES
    ('sqb', 'SQB', true),
    ('bitrix', 'Bitrix', true)
ON CONFLICT (code) DO NOTHING;

CREATE INDEX idx_integration_users_integration_id ON integration_users (integration_id);
CREATE INDEX idx_integration_users_user_id ON integration_users (user_id);
