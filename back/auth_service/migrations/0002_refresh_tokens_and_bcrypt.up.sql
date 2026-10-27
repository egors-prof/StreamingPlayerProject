-- Пароли теперь хранятся как bcrypt-хеш; старые SHA-256 хеши с bcrypt
-- несовместимы, такие пользователи должны зарегистрироваться заново.
ALTER TABLE users RENAME COLUMN password TO password_hash;

-- Выданные refresh-токены. Храним только идентификатор (jti), не сам токен.
CREATE TABLE IF NOT EXISTS refresh_tokens
(
    id         VARCHAR(64) PRIMARY KEY,
    user_id    INTEGER     NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    revoked_at TIMESTAMPTZ,
    -- Почему отозван: rotated (обменян на новый), logout, reuse_detected.
    -- Льготное окно для гонки вкладок действует только для rotated.
    revoked_reason VARCHAR(16),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_refresh_tokens_user_id ON refresh_tokens (user_id);
