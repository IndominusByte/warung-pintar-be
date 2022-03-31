CREATE TABLE IF NOT EXISTS account.confirmation_users(
  id SERIAL PRIMARY KEY,
  activated BOOLEAN DEFAULT false,
  resend_expired INT,
  user_id INT NOT NULL
);
