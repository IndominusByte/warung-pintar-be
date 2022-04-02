CREATE TABLE IF NOT EXISTS account.confirmation_users(
  id VARCHAR(100) DEFAULT uuid_in(overlay(overlay(md5(random()::text || ':' || clock_timestamp()::text) placing '4' from 13) placing to_hex(floor(random()*(11-8+1) + 8)::int)::text from 17)::cstring),
  activated BOOLEAN DEFAULT false,
  resend_expired TIMESTAMP WITHOUT TIME ZONE,
  user_id INT NOT NULL,
  PRIMARY KEY(id)
);
