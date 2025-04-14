CREATE TABLE users (
  id BIGSERIAL NOT NULL PRIMARY KEY,
  email VARCHAR(255) NOT NULL,
  hashed_password CHAR(60) NOT NULL,
  timezone VARCHAR(35) NOT NULL DEFAULT 'UTC',
  created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
  updated_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);

CREATE UNIQUE INDEX idx_users_email_unique ON users(email);

CREATE TABLE sessions (
  id BIGSERIAL PRIMARY KEY,
  user_id BIGINT NOT NULL REFERENCES users(id),
  token VARCHAR NOT NULL,
  ip_address VARCHAR NOT NULL,
  user_agent VARCHAR NOT NULL,
  expires_at TIMESTAMP NOT NULL,
  created_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC'),
  updated_at TIMESTAMP NOT NULL DEFAULT (NOW() AT TIME ZONE 'UTC')
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE UNIQUE INDEX idx_session_token_unique ON sessions(token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
