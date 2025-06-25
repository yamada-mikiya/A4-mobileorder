INSERT INTO users(email, role) VALUES
    ('admin1@example.com', 'admin'),
    ('admin2@example.com', 'admin'),
    ('admin3@example.com', 'admin');

INSERT INTO shops(name, description, location, admin_user_id) VALUES
    ('A4', 'おいしいの作ってます', 'A_35', 1),
    ('ラーメン屋', 'こだわっています', 'B_30', 2),
    ('定食屋', '種類多いです', 'B_31', 3);

INSERT INTO products (product_name, description, price) VALUES
    ('カステラ', 'こだわりカステラ', 2000),
    ('ラーメン', 'こだわりラーメン', 1000),
    ('おにぎり定食', 'こだわりおにぎり', 700);

INSERT INTO shop_product(shop_id, product_id) VALUES
    (1, 1),
    (2, 2),
    (3, 3);