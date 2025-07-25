-- データのクリア (開発時に毎回クリーンな状態にするため)
-- TRUNCATE TABLE users, shops, products, shop_product, orders, order_product RESTART IDENTITY CASCADE;

-- ユーザーを15人作成 (管理者5人、顧客10人)
-- role: 1 = Customer, 2 = Admin
INSERT INTO users(email, role) VALUES
('admin1@example.com', 2),
('admin2@example.com', 2),
('admin3@example.com', 2),
('admin4@example.com', 2),
('admin5@example.com', 2),
('customer1@example.com', 1),
('customer2@example.com', 1),
('customer3@example.com', 1),
('customer4@example.com', 1),
('customer5@example.com', 1),
('sato@example.com', 1),
('suzuki@example.com', 1),
('takahashi@example.com', 1),
('tanaka@example.com', 1),
('watanabe@example.com', 1);

-- 店舗を5件作成
INSERT INTO shops(name, description, location, admin_user_id) VALUES
('A4食堂', '安くて美味しい、学生街の定食屋です。', '神戸市灘区六甲台町1-1', 1),
('元町ラーメン一番星', '豚骨と魚介のWスープが自慢。', '神戸市中央区元町通2-9-1', 2),
('三宮ベーカリーカフェ', '毎朝焼き上げるパンと自家焙煎コーヒー。', '神戸市中央区三宮町1-8-1', 3),
('ハーバーランド・クレープ', '港の景色を眺めながら楽しむ、もちもちクレープ。', '神戸市中央区東川崎町1-6-1', 4),
('カフェ・ド・異人館', 'レトロな雰囲気でくつろぐ、北野の隠れ家カフェ。', '神戸市中央区北野町3-10-20', 5);

-- 商品を15件作成
INSERT INTO products (product_name, description, price) VALUES
-- 定食屋の商品 (ID: 1-4)
('唐揚げ定食', 'ジューシーなもも肉をカラッと揚げました。', 850),
('生姜焼き定食', '特製ダレが絡んだ豚ロース。ご飯が進みます。', 900),
('日替わりランチ', '毎日変わる、お得で栄養満点のランチセット。', 750),
('瓶ビール（中瓶）', '定食とご一緒にどうぞ。', 550),
-- ラーメン屋の商品 (ID: 5-7)
('特製豚骨ラーメン', '8時間煮込んだ濃厚スープ。', 950),
('味玉つけ麺', 'もちもちの太麺と濃厚魚介スープ。', 1050),
('チャーシュー丼', '特製ダレで煮込んだ絶品チャーシュー丼。', 400),
-- ベーカリーカフェの商品 (ID: 8-11)
('クロワッサン', 'フランス産発酵バターを贅沢に使用。', 320),
('バゲット', '外はカリカリ、中はもっちり。', 380),
('ブレンドコーヒー', '苦味と酸味のバランスが良いオリジナルブレンド。', 500),
('カフェラテ', 'エスプレッソとスチームミルクのハーモニー。', 580),
-- クレープ屋の商品 (ID: 12-13)
('チョコバナナ生クリーム', '王道の組み合わせ。一番人気です。', 600),
('ストロベリーチーズケーキ', '甘酸っぱい苺と濃厚チーズケーキ。', 750),
-- カフェの商品 (ID: 14-15)
('チーズケーキ', '濃厚でなめらかな口溶けのベイクドチーズケーキ。', 650),
('アイスティー', 'アールグレイの爽やかな香り。', 550);

-- 店舗と商品の関連付け (中間テーブル)
INSERT INTO shop_product(shop_id, product_id) VALUES
-- A4食堂 (shop_id: 1)
(1, 1), (1, 2), (1, 3), (1, 4),
-- 元町ラーメン一番星 (shop_id: 2)
(2, 5), (2, 6), (2, 7), (2, 4), -- ラーメン屋でもビールは売る
-- 三宮ベーカリーカフェ (shop_id: 3)
(3, 8), (3, 9), (3, 10), (3, 11),
-- ハーバーランド・クレープ (shop_id: 4)
(4, 12), (4, 13),
-- カフェ・ド・異人館 (shop_id: 5)
(5, 10), (5, 11), (5, 14), (5, 15); -- こちらのカフェでもコーヒーとラテは売る