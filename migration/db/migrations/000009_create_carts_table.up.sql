CREATE TABLE IF NOT EXISTS transaction.carts(
  id SERIAL PRIMARY KEY,
  notes VARCHAR(100),
  qty BIGINT NOT NULL,
  user_id INT NOT NULL,
  product_id INT NOT NULL
);
