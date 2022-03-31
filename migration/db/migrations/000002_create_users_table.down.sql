DROP TABLE IF EXISTS account.users;
DROP INDEX IF EXISTS idx_account_users_email on account.users(email);
DROP INDEX IF EXISTS idx_account_users_phone on account.users(phone);
