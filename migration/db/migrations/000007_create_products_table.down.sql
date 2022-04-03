DROP TABLE IF EXISTS product.products;
DROP INDEX IF EXISTS idx_product_products_name ON product.products(name);
DROP INDEX IF EXISTS idx_product_products_slug ON product.products(slug);
