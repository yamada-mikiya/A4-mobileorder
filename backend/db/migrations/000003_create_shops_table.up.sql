CREATE TABLE shops (
    shop_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    location VARCHAR(255),
    is_open BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trigger_update_shops_updated_at
BEFORE UPDATE ON shops
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();