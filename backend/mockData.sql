-- データのクリア (開発時に毎回クリーンな状態にするため)
-- 外部キー制約があるため、TRUNCATEの順番に注意
TRUNCATE TABLE order_item, orders, shop_staff, shop_item, users, shops, items RESTART IDENTITY CASCADE;

-- ユーザーを15人作成 (管理者5人、顧客10人)
-- role: 1 = Customer, 2 = Admin
INSERT INTO users(email, role) VALUES
('admin1@example.com', 2), -- ID: 1
('admin2@example.com', 2), -- ID: 2
('admin3@example.com', 2), -- ID: 3
('admin4@example.com', 2), -- ID: 4
('admin5@example.com', 2), -- ID: 5
('customer1@example.com', 1), -- ID: 6
('customer2@example.com', 1), -- ID: 7
('customer3@example.com', 1), -- ID: 8
('customer4@example.com', 1), -- ID: 9
('customer5@example.com', 1), -- ID: 10
('sato@example.com', 1),      -- ID: 11
('suzuki@example.com', 1),    -- ID: 12
('takahashi@example.com', 1), -- ID: 13
('tanaka@example.com', 1),    -- ID: 14
('watanabe@example.com', 1);  -- ID: 15

-- 店舗を5件作成
INSERT INTO shops(name, description, location) VALUES
('A4食堂', '安くて美味しい、学生街の定食屋です。', '神戸市灘区六甲台町1-1'),
('元町ラーメン一番星', '豚骨と魚介のWスープが自慢。', '神戸市中央区元町通2-9-1'),
('三宮ベーカリーカフェ', '毎朝焼き上げるパンと自家焙煎コーヒー。', '神戸市中央区三宮町1-8-1'),
('ハーバーランド・クレープ', '港の景色を眺めながら楽しむ、もちもちクレープ。', '神戸市中央区東川崎町1-6-1'),
('カフェ・ド・異人館', 'レトロな雰囲気でくつろぐ、北野の隠れ家カフェ。', '神戸市中央区北野町3-10-20');

-- 管理者と店舗の関連付け (shop_staffテーブル)
INSERT INTO shop_staff(user_id, shop_id) VALUES
(1, 1), -- admin1 は A4食堂 を担当
(2, 2), -- admin2 は 元町ラーメン一番星 を担当
(3, 3), -- admin3 は 三宮ベーカリーカフェ を担当
(4, 4), -- admin4 は ハーバーランド・クレープ を担当
(5, 5); -- admin5 は カフェ・ド・異人館 を担当

-- 商品を15件作成
INSERT INTO items (item_name, description, price) VALUES
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

-- 店舗と商品の関連付け (shop_itemテーブル)
INSERT INTO shop_item(shop_id, item_id) VALUES
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

-- 注文データ (ordersテーブル)
-- status: 1=cooking, 2=completed, 3=handed
INSERT INTO orders (user_id, shop_id, order_date, total_amount, guest_order_token, status) VALUES
-- customer1 (ID:6) の注文
(6, 1, NOW() - INTERVAL '20 minutes', 850, NULL, 1), -- A4食堂で唐揚げ定食 (調理中)
(6, 2, NOW() - INTERVAL '1 day', 1050, NULL, 2), -- 昨日、元町ラーメンでつけ麺 (調理完了)
-- customer2 (ID:7) の注文
(7, 1, NOW() - INTERVAL '10 minutes', 900, NULL, 1), -- A4食堂で生姜焼き定食 (調理中)
-- ゲストユーザーの注文 (user_idがNULL)
(NULL, 3, NOW() - INTERVAL '5 minutes', 900, 'guest-token-12345', 1); -- 三宮ベーカリーでクロワッサンとコーヒー (調理中)

-- 注文と商品の関連付け (order_itemテーブル)
INSERT INTO order_item (order_id, item_id, quantity, price_at_order) VALUES
-- 注文ID: 1 (唐揚げ定食)
(1, 1, 1, 850),
-- 注文ID: 2 (味玉つけ麺)
(2, 6, 1, 1050),
-- 注文ID: 3 (生姜焼き定食)
(3, 2, 1, 900),
-- 注文ID: 4 (クロワッサンとコーヒー)
(4, 8, 1, 320),
(4, 10, 1, 500);
