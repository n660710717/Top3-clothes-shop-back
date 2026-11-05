# Clothes Shop

ร้านเสื้อผ้ามีสองโปรเจกต์ คนละ git repo

- หลังบ้าน `Top3-clothes-shop-back` เขียนด้วย Go ตามแนวแล็บ week6 (handler / store / mw / auth, ตั๋ว JWT, PostgreSQL)
- หน้าบ้าน `Top3-clothes-shop-front` เป็น React ที่พอร์ต 3000 ยิงมาที่ `http://localhost:8080`

รหัสสมาชิกทุกจุดมาจากตั๋วเข้าสู่ระบบ หน้าเว็บส่งชื่อหรือรหัสคนอื่นมาเองไม่ได้

## บัญชีทดลอง

| อีเมล | รหัสผ่าน | บทบาท |
| --- | --- | --- |
| admin@clothinghub.com | Admin1234! | admin |
| somchai@seller.com | Seller1234! | seller |
| buyer@store.com | Buyer1234! | buyer |

สินค้า 150 ชิ้นเดิมเป็นของร้าน somchai (สมาชิก id 2)

## รัน

1. ที่โฟลเดอร์นี้: `docker compose up -d`
2. ถ้าฐานถูกสร้างไว้ก่อนหน้าตารางด้านล่าง ให้อัปเดตครั้งเดียว

```powershell
Get-Content .\clotesShop\user_scope.sql | docker exec -i top3-clothes-shop-back-db-1 psql -U postgres -d clothesshop
```

3. ที่ `clotesShop`: `go run .` ฟังพอร์ต 8080
4. ที่หน้าบ้าน: มีไฟล์ `.env` ค่า `REACT_APP_API_URL=http://localhost:8080` แล้ว `npm start`

ฐานที่สร้างใหม่จะรัน `schema.sql`, `seed.sql`, `images.sql`, `members.sql`, `user_scope.sql` ตามลำดับเอง

## ข้อมูลที่แยกตามคน

ก่อนหน้านี้ถูกใจ ตะกร้า คำสั่งซื้อ และคำร้องเรียนถูกเก็บใน `localStorage` ก้อนเดียวของเบราว์เซอร์ รีวิวอยู่แค่ในหน่วยความจำของหน้าเว็บ คนต่อไปที่เข้าเครื่องเดียวกันจึงเห็นของคนก่อน และคนละเครื่องไม่เห็นคอมเมนต์เดียวกัน โปรไฟล์ยังโชว์คำสั่งซื้อปลอมของ Praewa Buyer แม้ไม่ได้เข้าสู่ระบบ

ตอนนี้เก็บใน PostgreSQL และผูก `member_id` จากตั๋ว

| ตาราง | ใครเห็น | หมายเหตุ |
| --- | --- | --- |
| wishlists | เจ้าของบัญชีผู้ซื้อ | หัวใจไม่ปนกันเมื่อสลับบัญชี |
| carts | เจ้าของบัญชีผู้ซื้อ | แยกตามสินค้าและไซซ์ |
| orders, order_items | ผู้ซื้อเห็นออเดอร์ตัวเอง ผู้ขายเห็นเฉพาะสินค้าของร้านตัวเอง | กดสั่งซื้อจากตะกร้าแล้วตะกร้าใบนั้นว่าง |
| reviews | ทุกคนที่เปิดสินค้า | ชื่อผู้เขียนมาจากตาราง members ไม่ใช่ชื่อที่หน้าเว็บส่งมา ผู้ซื้อหนึ่งคนรีวิวสินค้าหนึ่งชิ้นได้ครั้งเดียว แอดมินลบได้ |
| disputes | ผู้ซื้อเป็นคนแจ้ง แอดมินเห็นทุกเรื่องที่ยังไม่ปิด | ไม่ผูกกับเครื่องที่เปิดเบราว์เซอร์ |
| chat_messages | เฉพาะคู่ผู้ซื้อคนนั้นกับสินค้าชิ้นนั้น | คนซื้อคนละบัญชีไม่เห็นแชทของกันและกัน |
| products.seller_id | ผู้ขายเห็นและแก้ได้เฉพาะสินค้าของตัวเอง | ลบหรือแก้ของร้านอื่นไม่ได้ |

## API ที่เพิ่ม

สาธารณะ

- `GET /api/products/{id}/reviews` ตอบ `{ items: [{ id, user, rating, comment, date }] }`

ผู้ซื้อ ต้องมีตั๋วและบทบาท buyer

- `GET /api/me/wishlist` และ `PUT /api/me/wishlist/{productId}` สลับถูกใจ ตอบ `{ liked }`
- `GET /api/me/cart`
- `POST /api/me/cart` เพิ่มจำนวน ส่ง `{ productId, size, quantity }`
- `PUT /api/me/cart` กำหนดจำนวน ถ้า quantity เป็น 0 คือลบออก
- `POST /api/me/orders` สั่งซื้อจากตะกร้า
- `GET /api/me/orders` ประวัติของคนนั้น
- `POST /api/products/{id}/reviews` ส่ง `{ rating, comment }`
- `GET` และ `POST /api/products/{id}/messages`
- `POST /api/products/{id}/disputes` ส่ง `{ reason }`

ผู้ขาย

- `GET /api/seller/products` เฉพาะของร้านนี้
- `POST /api/seller/products` และ `PUT /api/seller/products/{id}`
- `DELETE /api/seller/products/{id}` ลบได้เฉพาะของตัวเอง
- `GET /api/seller/orders`
- `PUT /api/seller/orders/{id}` ส่ง `{ trackingNo }` แล้วสถานะเป็นจัดส่งแล้ว

แอดมิน

- `GET /api/admin/reviews` และ `DELETE /api/reviews/{id}`
- `GET /api/admin/disputes` และ `DELETE /api/admin/disputes/{id}` ปิดเรื่อง

ของเดิมยังอยู่: สมัคร เข้าสู่ระบบ เปลี่ยนรหัสผ่าน รายการสินค้า สมาชิก และเปลี่ยนบทบาท

## สิ่งที่หน้าบ้านเปลี่ยน

ไฟล์หลักคือ `src/api/backend.js` แล้วหน้าเหล่านี้เลิกใช้กุญแจ `wishlist`, `cart`, `disputes`, `seller_orders` ในเบราว์เซอร์

- การ์ดสินค้าและหน้ารายการที่ชอบ บันทึกหัวใจของบัญชีที่เข้าอยู่
- ตะกร้าและปุ่มสั่งซื้อ
- หน้ารายละเอียดสินค้า: รีวิวที่ทุกคนเห็นพร้อมชื่อผู้เขียน, แชทของบัญชีนั้น, ร้องเรียนเข้าแอดมิน, ใส่ตะกร้าตามไซซ์
- โปรไฟล์ไม่แสดงชื่อ Praewa Buyer ปลอม และคำสั่งซื้อมาจากฐานข้อมูล
- แดชบอร์ดผู้ขายโหลดเฉพาะสินค้าและออเดอร์ของร้าน เพิ่ม แก้ ลบ และใส่เลขพัสดุลงฐานข้อมูล
- แดชบอร์ดแอดมินโหลดคำร้องและรีวิวจริง ลบรีวิวหรือปิดคำร้องได้
- ออกจากระบบแล้วตัวนับหัวใจกับตะกร้ากลับเป็นศูนย์ เพราะไปอ่านของบัญชีใหม่

คำขอเปิดร้านในแท็บอนุมัติผู้ขายยังเป็นตัวอย่างบนหน้าเว็บ ไม่ได้ผูกบัญชีสมาชิก
