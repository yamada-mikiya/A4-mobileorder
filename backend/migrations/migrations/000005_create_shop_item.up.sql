CREATE TABLE shop_product (
    shop_product_id SERIAL PRIMARY KEY,
    shop_id INT NOT NULL,
    product_id INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (shop_id, product_id),
    FOREIGN KEY (shop_id) REFERENCES shops(shop_id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(product_id) ON DELETE CASCADE
);

CREATE TRIGGER trigger_update_shop_product_updated_at
BEFORE UPDATE ON shop_product
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();