package store

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type SellerApplication struct {
	ID       int64  `json:"id"`
	ShopName string `json:"shopName"`
	Owner    string `json:"owner"`
	Email    string `json:"email"`
	Date     string `json:"date"`
}

func (s *Store) CreateSellerApplication(ctx context.Context, memberID int64, shopName string) error {
	_, err := s.Pool.Exec(ctx, `
INSERT INTO seller_applications (member_id, shop_name)
VALUES ($1, $2)`, memberID, shopName)
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrAlreadyApplied
	}
	return err
}

func (s *Store) ListSellerApplications(ctx context.Context) ([]SellerApplication, error) {
	rows, err := s.Pool.Query(ctx, `
SELECT a.id, a.shop_name, m.display_name, m.email, a.created_at
  FROM seller_applications a
  JOIN members m ON m.id = a.member_id
 WHERE a.status = 'pending'
 ORDER BY a.id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SellerApplication{}
	for rows.Next() {
		var item SellerApplication
		var at time.Time
		if err := rows.Scan(&item.ID, &item.ShopName, &item.Owner, &item.Email, &at); err != nil {
			return nil, err
		}
		item.Date = dayStamp(at)
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) ApproveSellerApplication(ctx context.Context, id int64) error {
	tx, err := s.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	var memberID int64
	err = tx.QueryRow(ctx, `
SELECT member_id FROM seller_applications
 WHERE id = $1 AND status = 'pending'
 FOR UPDATE`, id).Scan(&memberID)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
UPDATE members SET role = 'seller', updated_at = now() WHERE id = $1`, memberID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `
UPDATE seller_applications SET status = 'approved' WHERE id = $1`, id); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
