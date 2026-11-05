package store

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var (
	ErrAlreadyReviewed = errors.New("already reviewed")
	ErrEmptyCart       = errors.New("empty cart")
)

type AdminReview struct {
	ID          int64  `json:"id"`
	ProductName string `json:"productName"`
	User        string `json:"user"`
	Comment     string `json:"comment"`
}

type Review struct {
	ID      int64  `json:"id"`
	User    string `json:"user"`
	Rating  int    `json:"rating"`
	Comment string `json:"comment"`
	Date    string `json:"date"`
}

type CartItem struct {
	ID         int64  `json:"id"`
	Name       string `json:"name"`
	Price      int    `json:"price"`
	Image      string `json:"image"`
	ChosenSize string `json:"chosenSize"`
	Quantity   int    `json:"quantity"`
}

type BuyerOrder struct {
	ID     int64  `json:"id"`
	Date   string `json:"date"`
	Total  int    `json:"total"`
	Status string `json:"status"`
	Items  string `json:"items"`
}

type SellerOrder struct {
	ID         int64  `json:"id"`
	Customer   string `json:"customer"`
	Item       string `json:"item"`
	Amount     int    `json:"amount"`
	Status     string `json:"status"`
	TrackingNo string `json:"trackingNo"`
}

type Dispute struct {
	ID          int64  `json:"id"`
	ProductName string `json:"productName"`
	BuyerName   string `json:"buyerName"`
	Reason      string `json:"reason"`
	Date        string `json:"date"`
}

type ChatMessage struct {
	ID     int64  `json:"id"`
	Sender string `json:"sender"`
	Text   string `json:"text"`
	Time   string `json:"time"`
}

type ProductInput struct {
	Name     string
	Category string
	Price    int
	Stock    int
	Image    string
}

func shopTime() *time.Location {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		return time.Local
	}
	return loc
}

func dayStamp(t time.Time) string {
	return t.In(shopTime()).Format("2006-01-02")
}

func clockStamp(t time.Time) string {
	return t.In(shopTime()).Format("15:04")
}

func (s *Store) productExists(ctx context.Context, id int64) error {
	var one int
	err := s.Pool.QueryRow(ctx, `SELECT 1 FROM products WHERE id = $1`, id).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func (s *Store) ListReviews(ctx context.Context, productID int64) ([]Review, error) {
	rows, err := s.Pool.Query(ctx, `
SELECT r.id, m.display_name, r.rating, r.comment, r.created_at
  FROM reviews r
  JOIN members m ON m.id = r.member_id
 WHERE r.product_id = $1
 ORDER BY r.created_at DESC`, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Review{}
	for rows.Next() {
		var item Review
		var at time.Time
		if err := rows.Scan(&item.ID, &item.User, &item.Rating, &item.Comment, &at); err != nil {
			return nil, err
		}
		item.Date = dayStamp(at)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) CreateReview(ctx context.Context, productID, memberID int64, rating int, comment string) error {
	if err := s.productExists(ctx, productID); err != nil {
		return err
	}
	_, err := s.Pool.Exec(ctx, `
INSERT INTO reviews (product_id, member_id, rating, comment)
VALUES ($1, $2, $3, $4)`, productID, memberID, rating, comment)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrAlreadyReviewed
	}
	return err
}

func (s *Store) ListAllReviews(ctx context.Context) ([]AdminReview, error) {
	rows, err := s.Pool.Query(ctx, `
SELECT r.id, p.name, m.display_name, r.comment
  FROM reviews r
  JOIN products p ON p.id = r.product_id
  JOIN members m ON m.id = r.member_id
 ORDER BY r.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []AdminReview{}
	for rows.Next() {
		var item AdminReview
		if err := rows.Scan(&item.ID, &item.ProductName, &item.User, &item.Comment); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) DeleteReview(ctx context.Context, id int64) error {
	tag, err := s.Pool.Exec(ctx, `DELETE FROM reviews WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListWishlist(ctx context.Context, memberID int64) ([]Product, error) {
	rows, err := s.Pool.Query(ctx, productSelect+`
  JOIN wishlists w ON w.product_id = p.id
 WHERE w.member_id = $1
 GROUP BY p.id, i.image_url, w.created_at
 ORDER BY w.created_at DESC`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Product{}
	for rows.Next() {
		p, err := scanProduct(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) ToggleWishlist(ctx context.Context, memberID, productID int64) (bool, error) {
	if err := s.productExists(ctx, productID); err != nil {
		return false, err
	}
	tag, err := s.Pool.Exec(ctx, `
DELETE FROM wishlists WHERE member_id = $1 AND product_id = $2`, memberID, productID)
	if err != nil {
		return false, err
	}
	if tag.RowsAffected() > 0 {
		return false, nil
	}
	_, err = s.Pool.Exec(ctx, `
INSERT INTO wishlists (member_id, product_id) VALUES ($1, $2)`, memberID, productID)
	return err == nil, err
}

func (s *Store) ListCart(ctx context.Context, memberID int64) ([]CartItem, error) {
	rows, err := s.Pool.Query(ctx, `
SELECT p.id, p.name, p.price, COALESCE(i.image_url, ''), c.size, c.quantity
  FROM carts c
  JOIN products p ON p.id = c.product_id
  LEFT JOIN product_images i ON i.product_id = p.id
 WHERE c.member_id = $1
 ORDER BY p.id, c.size`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []CartItem{}
	for rows.Next() {
		var item CartItem
		if err := rows.Scan(&item.ID, &item.Name, &item.Price, &item.Image, &item.ChosenSize, &item.Quantity); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) AddCart(ctx context.Context, memberID, productID int64, size string, qty int) error {
	if err := s.productExists(ctx, productID); err != nil {
		return err
	}
	_, err := s.Pool.Exec(ctx, `
INSERT INTO carts (member_id, product_id, size, quantity)
VALUES ($1, $2, $3, $4)
ON CONFLICT (member_id, product_id, size)
DO UPDATE SET quantity = carts.quantity + EXCLUDED.quantity`, memberID, productID, size, qty)
	return err
}

func (s *Store) SetCart(ctx context.Context, memberID, productID int64, size string, qty int) error {
	if qty <= 0 {
		_, err := s.Pool.Exec(ctx, `
DELETE FROM carts WHERE member_id = $1 AND product_id = $2 AND size = $3`, memberID, productID, size)
		return err
	}
	tag, err := s.Pool.Exec(ctx, `
UPDATE carts SET quantity = $4
 WHERE member_id = $1 AND product_id = $2 AND size = $3`, memberID, productID, size, qty)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	return s.AddCart(ctx, memberID, productID, size, qty)
}

func (s *Store) Checkout(ctx context.Context, buyerID int64) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	rows, err := tx.Query(ctx, `
SELECT c.product_id, p.name, c.size, c.quantity, p.price, COALESCE(p.seller_id, 2)
  FROM carts c
  JOIN products p ON p.id = c.product_id
 WHERE c.member_id = $1`, buyerID)
	if err != nil {
		return err
	}
	type line struct {
		productID int64
		name      string
		size      string
		qty       int
		price     int
		sellerID  int64
	}
	lines := []line{}
	total := 0
	for rows.Next() {
		var item line
		if err := rows.Scan(&item.productID, &item.name, &item.size, &item.qty, &item.price, &item.sellerID); err != nil {
			rows.Close()
			return err
		}
		total += item.price * item.qty
		lines = append(lines, item)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	if len(lines) == 0 {
		return ErrEmptyCart
	}
	var orderID int64
	if err := tx.QueryRow(ctx, `
INSERT INTO orders (buyer_id, total) VALUES ($1, $2) RETURNING id`, buyerID, total).Scan(&orderID); err != nil {
		return err
	}
	for _, item := range lines {
		if _, err := tx.Exec(ctx, `
INSERT INTO order_items
    (order_id, product_id, seller_id, product_name, size, quantity, unit_price)
VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			orderID, item.productID, item.sellerID, item.name, item.size, item.qty, item.price); err != nil {
			return err
		}
	}
	if _, err := tx.Exec(ctx, `DELETE FROM carts WHERE member_id = $1`, buyerID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) ListBuyerOrders(ctx context.Context, buyerID int64) ([]BuyerOrder, error) {
	rows, err := s.Pool.Query(ctx, `
SELECT o.id, o.created_at, o.total,
       string_agg(oi.product_name || ' x' || oi.quantity::text, ', ' ORDER BY oi.id),
       CASE WHEN bool_and(oi.status = 'จัดส่งแล้ว') THEN 'จัดส่งแล้ว' ELSE 'รอการจัดส่ง' END
  FROM orders o
  JOIN order_items oi ON oi.order_id = o.id
 WHERE o.buyer_id = $1
 GROUP BY o.id
 ORDER BY o.id DESC`, buyerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []BuyerOrder{}
	for rows.Next() {
		var item BuyerOrder
		var at time.Time
		if err := rows.Scan(&item.ID, &at, &item.Total, &item.Items, &item.Status); err != nil {
			return nil, err
		}
		item.Date = dayStamp(at)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) ListSellerOrders(ctx context.Context, sellerID int64) ([]SellerOrder, error) {
	rows, err := s.Pool.Query(ctx, `
SELECT oi.id, m.display_name,
       oi.product_name || ' (ไซซ์ ' || oi.size || ') × ' || oi.quantity::text,
       oi.unit_price * oi.quantity, oi.status, oi.tracking_no
  FROM order_items oi
  JOIN orders o ON o.id = oi.order_id
  JOIN members m ON m.id = o.buyer_id
 WHERE oi.seller_id = $1
 ORDER BY oi.id DESC`, sellerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SellerOrder{}
	for rows.Next() {
		var item SellerOrder
		if err := rows.Scan(&item.ID, &item.Customer, &item.Item, &item.Amount, &item.Status, &item.TrackingNo); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) ShipOrder(ctx context.Context, itemID, sellerID int64, tracking string) error {
	tag, err := s.Pool.Exec(ctx, `
UPDATE order_items
   SET status = 'จัดส่งแล้ว', tracking_no = $3
 WHERE id = $1 AND seller_id = $2`, itemID, sellerID, tracking)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) CreateDispute(ctx context.Context, productID, buyerID int64, reason string) error {
	if err := s.productExists(ctx, productID); err != nil {
		return err
	}
	_, err := s.Pool.Exec(ctx, `
INSERT INTO disputes (product_id, buyer_id, reason) VALUES ($1, $2, $3)`, productID, buyerID, reason)
	return err
}

func (s *Store) ListDisputes(ctx context.Context) ([]Dispute, error) {
	rows, err := s.Pool.Query(ctx, `
SELECT d.id, COALESCE(p.name, 'สินค้าที่ถูกลบ'), m.display_name, d.reason, d.created_at
  FROM disputes d
  JOIN members m ON m.id = d.buyer_id
  LEFT JOIN products p ON p.id = d.product_id
 WHERE NOT d.resolved
 ORDER BY d.id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Dispute{}
	for rows.Next() {
		var item Dispute
		var at time.Time
		if err := rows.Scan(&item.ID, &item.ProductName, &item.BuyerName, &item.Reason, &at); err != nil {
			return nil, err
		}
		item.Date = dayStamp(at)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) ResolveDispute(ctx context.Context, id int64) error {
	tag, err := s.Pool.Exec(ctx, `UPDATE disputes SET resolved = TRUE WHERE id = $1 AND NOT resolved`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *Store) ListMessages(ctx context.Context, productID, buyerID int64) ([]ChatMessage, error) {
	rows, err := s.Pool.Query(ctx, `
SELECT id, sender, body, created_at
  FROM chat_messages
 WHERE product_id = $1 AND buyer_id = $2
 ORDER BY created_at, id`, productID, buyerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []ChatMessage{}
	for rows.Next() {
		var item ChatMessage
		var at time.Time
		if err := rows.Scan(&item.ID, &item.Sender, &item.Text, &at); err != nil {
			return nil, err
		}
		item.Time = clockStamp(at)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) AddMessage(ctx context.Context, productID, buyerID int64, text string) error {
	if err := s.productExists(ctx, productID); err != nil {
		return err
	}
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `
INSERT INTO chat_messages (product_id, buyer_id, sender, body)
VALUES ($1, $2, 'user', $3)`, productID, buyerID, text); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO chat_messages (product_id, buyer_id, sender, body)
VALUES ($1, $2, 'seller', 'ได้รับข้อความแล้วค่ะ ร้านจะตอบกลับโดยเร็ว')`, productID, buyerID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) CreateProduct(ctx context.Context, sellerID int64, in ProductInput) (Product, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Product{}, err
	}
	defer tx.Rollback(ctx)
	var id int64
	err = tx.QueryRow(ctx, `
INSERT INTO products (id, name, title, category, price, stock, description, seller_id)
SELECT COALESCE(MAX(id), 0) + 1, $1, $1, $2, $3, $4, $5, $6 FROM products
RETURNING id`, in.Name, in.Category, in.Price, in.Stock, in.Name, sellerID).Scan(&id)
	if err != nil {
		return Product{}, err
	}
	for i, size := range []string{"S", "M", "L"} {
		if _, err := tx.Exec(ctx, `
INSERT INTO product_sizes (product_id, size, position) VALUES ($1, $2, $3)`, id, size, i+1); err != nil {
			return Product{}, err
		}
	}
	if strings.TrimSpace(in.Image) != "" {
		if _, err := tx.Exec(ctx, `
INSERT INTO product_images (product_id, image_url) VALUES ($1, $2)`, id, strings.TrimSpace(in.Image)); err != nil {
			return Product{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Product{}, err
	}
	return s.ProductByID(ctx, id)
}

func (s *Store) UpdateProduct(ctx context.Context, id, sellerID int64, in ProductInput) (Product, error) {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return Product{}, err
	}
	defer tx.Rollback(ctx)
	tag, err := tx.Exec(ctx, `
UPDATE products
   SET name = $3, title = $3, category = $4, price = $5, stock = $6, description = $3
 WHERE id = $1 AND seller_id = $2`, id, sellerID, in.Name, in.Category, in.Price, in.Stock)
	if err != nil {
		return Product{}, err
	}
	if tag.RowsAffected() == 0 {
		return Product{}, ErrNotFound
	}
	if strings.TrimSpace(in.Image) != "" {
		if _, err := tx.Exec(ctx, `
INSERT INTO product_images (product_id, image_url) VALUES ($1, $2)
ON CONFLICT (product_id) DO UPDATE SET image_url = EXCLUDED.image_url`, id, strings.TrimSpace(in.Image)); err != nil {
			return Product{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return Product{}, err
	}
	return s.ProductByID(ctx, id)
}
