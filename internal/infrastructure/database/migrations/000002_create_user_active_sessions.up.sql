CREATE TABLE user_active_sessions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    device_id VARCHAR(255) NOT NULL,
    refresh_token_hash VARCHAR(255) NOT NULL,
    last_login_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_user_active_sessions_user_id ON user_active_sessions (user_id);
CREATE INDEX idx_user_active_sessions_refresh_token_hash ON user_active_sessions (refresh_token_hash);
CREATE INDEX idx_user_active_sessions_device_id ON user_active_sessions (user_id, device_id);
