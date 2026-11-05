package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type Product struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Title       string   `json:"title"`
	Category    string   `json:"category"`
	Price       int      `json:"price"`
	Stock       int      `json:"stock"`
	Sizes       []string `json:"sizes"`
	Description string   `json:"description"`
	Image       string   `json:"image"`
	ImageURL    string   `json:"image_url"`
}

const productSelect = `
SELECT p.id, p.name, p.title, p.category, p.price, p.stock, p.description,
       COALESCE(i.image_url, ''),
       COALESCE(array_agg(s.size ORDER BY s.position) FILTER (WHERE s.size IS NOT NULL), '{}')
  FROM products p
  LEFT JOIN product_images i ON i.product_id = p.id
  LEFT JOIN product_sizes s ON s.product_id = p.id`

func scanProduct(scan func(dest ...any) error) (Product, error) {
	var p Product
	err := scan(&p.ID, &p.Name, &p.Title, &p.Category, &p.Price, &p.Stock, &p.Description, &p.Image, &p.Sizes)
	if p.Sizes == nil {
		p.Sizes = []string{}
	}
	p.ImageURL = p.Image
	return p, err
}

func (s *Store) ListProducts(ctx context.Context) ([]Product, error) {
	rows, err := s.Pool.Query(ctx, productSelect+`
 GROUP BY p.id, i.image_url
 ORDER BY p.id`)
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

func (s *Store) ProductByID(ctx context.Context, id int64) (Product, error) {
	row := s.Pool.QueryRow(ctx, productSelect+`
 WHERE p.id = $1
 GROUP BY p.id, i.image_url`, id)
	p, err := scanProduct(row.Scan)
	if errors.Is(err, pgx.ErrNoRows) {
		return Product{}, ErrNotFound
	}
	return p, err
}

func (s *Store) ListProductsBySeller(ctx context.Context, sellerID int64) ([]Product, error) {
	rows, err := s.Pool.Query(ctx, productSelect+`
 WHERE p.seller_id = $1
 GROUP BY p.id, i.image_url
 ORDER BY p.id`, sellerID)
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

func (s *Store) DeleteProduct(ctx context.Context, id, sellerID int64) error {
	tag, err := s.Pool.Exec(ctx, `DELETE FROM products WHERE id = $1 AND seller_id = $2`, id, sellerID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
