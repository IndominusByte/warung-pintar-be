CREATE TABLE IF NOT EXISTS account.users(
  id SERIAL PRIMARY KEY,
  fullname VARCHAR(100),
  email VARCHAR(100) UNIQUE NOT NULL,
  password VARCHAR(100) NOT NULL,
  phone VARCHAR(20) UNIQUE,
  address TEXT,
  role VARCHAR(10) DEFAULT 'guest',
  avatar VARCHAR(100) DEFAULT 'default.jpg',
  created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_account_users_email on account.users(email);
CREATE INDEX IF NOT EXISTS idx_account_users_phone on account.users(phone);
