CREATE TABLE IF NOT EXISTS product.products(
  id SERIAL PRIMARY KEY,
  name VARCHAR(100) UNIQUE NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  description TEXT NOT NULL,
  image VARCHAR(100) NOT NULL,
  price BIGINT NOT NULL,
  stock BIGINT NOT NULL,
  category_id INT NOT NULL,
  created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_product_products_name ON product.products(name);
CREATE INDEX IF NOT EXISTS idx_product_products_slug ON product.products(slug);
