CREATE TABLE IF NOT EXISTS account.password_resets(
  id SERIAL PRIMARY KEY,
  email VARCHAR(100) UNIQUE NOT NULL,
  resend_expired INT,
  created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);


CREATE INDEX IF NOT EXISTS idx_account_password_resets_email on account.password_resets(email);
