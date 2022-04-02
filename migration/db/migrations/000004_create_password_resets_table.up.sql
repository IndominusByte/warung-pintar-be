CREATE TABLE IF NOT EXISTS account.password_resets(
  id VARCHAR(100) DEFAULT uuid_in(overlay(overlay(md5(random()::text || ':' || clock_timestamp()::text) placing '4' from 13) placing to_hex(floor(random()*(11-8+1) + 8)::int)::text from 17)::cstring),
  email VARCHAR(100) UNIQUE NOT NULL,
  resend_expired TIMESTAMP WITHOUT TIME ZONE,
  created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY(id)
);

CREATE INDEX IF NOT EXISTS idx_account_password_resets_email on account.password_resets(email);
