CREATE TABLE orders (
    order_id SERIAL PRIMARY KEY,
    user_id INT,
    shop_id INT NOT NULL,
    order_date TIMESTAMP NOT NULL,
    total_amount INTEGER NOT NULL,
    guest_order_token VARCHAR(255) UNIQUE NULL,
    status SMALLINT NOT NULL DEFAULT 1, -- 1: Cooking, 2: Completed, 3: Handed
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(user_id),
    FOREIGN KEY (shop_id) REFERENCES shops(shop_id)
);

CREATE TRIGGER trigger_update_orders_updated_at
BEFORE UPDATE ON orders
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();