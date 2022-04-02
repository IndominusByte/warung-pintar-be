DROP TABLE IF EXISTS account.password_resets;
DROP INDEX IF EXISTS idx_account_password_resets_email on account.password_resets(email);
