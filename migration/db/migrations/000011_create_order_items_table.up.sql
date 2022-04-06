CREATE TABLE IF NOT EXISTS transaction.order_items(
  id SERIAL PRIMARY KEY,
  notes VARCHAR(100),
  qty BIGINT NOT NULL,
  price BIGINT NOT NULL,
  product_id INT NOT NULL,
  order_id INT NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);