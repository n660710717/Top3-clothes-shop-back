-- บัญชีเริ่มต้นให้ตรงกับหน้าร้าน
-- รัน schema.sql ก่อนไฟล์นี้
-- admin@clothinghub.com  รหัส Admin1234!
-- somchai@seller.com     รหัส Seller1234!
-- buyer@store.com        รหัส Buyer1234!

BEGIN;

INSERT INTO members (id, email, password_hash, display_name, role) VALUES
(1, 'admin@clothinghub.com', '$2b$10$Wdu8Lx3f.J08TYv4XgMvW.VltDKqSHIctNL6Ab4Fc3NiwcCclzfsy', 'Admin Manager', 'admin'),
(2, 'somchai@seller.com', '$2b$10$BEhKb8uamtOWRpaOg3632.0WNv8AuW3hvLpyHpjUadMg2ZmIfq5vm', 'Somchai Shop', 'seller'),
(3, 'buyer@store.com', '$2b$10$bWWGYk/6fIyNtZVbCyY7mejD0e1j06GmE1WzRRfyp1/1Jr3KDUkpa', 'Praewa Rueangsri', 'buyer');

SELECT setval(pg_get_serial_sequence('members', 'id'), (SELECT MAX(id) FROM members));

COMMIT;
