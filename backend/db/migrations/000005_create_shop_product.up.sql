CREATE TABLE shop_item (
    shop_item_id SERIAL PRIMARY KEY,
    shop_id INT NOT NULL,
    item_id INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE (shop_id, item_id),
    FOREIGN KEY (shop_id) REFERENCES shops(shop_id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items(item_id) ON DELETE CASCADE
);

CREATE TRIGGER trigger_update_shop_item_updated_at
BEFORE UPDATE ON shop_item
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();