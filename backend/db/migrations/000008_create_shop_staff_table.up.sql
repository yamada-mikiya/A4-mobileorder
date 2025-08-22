CREATE TABLE shop_staff (
    shop_staff_id SERIAL PRIMARY KEY,
    shop_id INT NOT NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (shop_id) REFERENCES shops(shop_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    UNIQUE (shop_id, user_id) -- 同じユーザーが同じ店に複数登録されるのを防ぐ
);

CREATE TRIGGER trigger_shop_staff_updated_at
BEFORE UPDATE ON shop_staff
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();