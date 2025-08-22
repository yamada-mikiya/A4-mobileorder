DROP TRIGGER IF EXISTS trigger_shop_staff_updated_at ON shop_staff;
DROP TABLE IF EXISTS shop_staff;

DROP TRIGGER IF EXISTS trigger_update_orders_updated_at ON order_item;
DROP TABLE IF EXISTS order_item;

DROP TRIGGER IF EXISTS trigger_update_orders_updated_at ON orders;
DROP TABLE IF EXISTS orders;

DROP TRIGGER IF EXISTS trigger_update_shop_item_updated_at ON shop_item;
DROP TABLE IF EXISTS shop_item;

DROP TRIGGER IF EXISTS trigger_update_items_updated_at ON items;
DROP TABLE IF EXISTS items;

DROP TRIGGER IF EXISTS trigger_update_shops_updated_at ON shops;
DROP TABLE IF EXISTS shops;

DROP TRIGGER IF EXISTS trigger_update_users_updated_at ON users;
DROP TABLE IF EXISTS users;

DROP FUNCTION IF EXISTS update_updated_at_column;


-- 000001_create_update_function.up.sql
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW(); -- 'NEW'は更新後の行データを表す
   RETURN NEW;
END;
$$ language 'plpgsql';

-- 000002_create_users_table.up.sql
CREATE TABLE users (
    user_id SERIAL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    role SMALLINT NOT NULL DEFAULT 1, -- 1: CustomerRole, 2: AdminRole
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trigger_update_users_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- 000003_create_shops_table.up.sql
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

-- 000004_create_products_table.up.sql
CREATE TABLE items (
    item_id SERIAL PRIMARY KEY,
    item_name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    is_available BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TRIGGER trigger_update_items_updated_at
BEFORE UPDATE ON items
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- 000005_create_shop_product.up.sql
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

-- 000006_create_orders_table.up.sql
CREATE TABLE orders (
    order_id SERIAL PRIMARY KEY,
    user_id INT,
    shop_id INT NOT NULL,
    order_date TIMESTAMP NOT NULL,
    total_amount DECIMAL(10, 2) NOT NULL,
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

-- 000007_create_order_item_table.up.sql
CREATE TABLE order_item (
    order_item_id SERIAL PRIMARY KEY,
    order_id INT NOT NULL,
    item_id INT NOT NULL,
    quantity INT NOT NULL,
    price_at_order DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (order_id) REFERENCES orders(order_id) ON DELETE CASCADE,
    FOREIGN KEY (item_id) REFERENCES items(item_id)
);

CREATE TRIGGER trigger_update_orders_updated_at
BEFORE UPDATE ON order_item
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();

-- 000008_create_shop_staff_table.up.sql
CREATE TABLE shop_staff (
    shop_staff_id SERIAL PRIMARY KEY,
    shop_id INT NOT NULL,
    user_id INT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    FOREIGN KEY (shop_id) REFERENCES shops(shop_id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(user_id) ON DELETE CASCADE,
    UNIQUE (shop_id, user_id)
);

CREATE TRIGGER trigger_shop_staff_updated_at
BEFORE UPDATE ON shop_staff
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();