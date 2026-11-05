package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

func (s *Store) SetMemberRole(ctx context.Context, id int64, role string) (Member, error) {
	var m Member
	err := s.Pool.QueryRow(ctx,
		`UPDATE members
            SET role = $2, updated_at = now()
          WHERE id = $1
         RETURNING id, email, display_name, role`, id, role).
		Scan(&m.ID, &m.Email, &m.DisplayName, &m.Role)
	if errors.Is(err, pgx.ErrNoRows) {
		return Member{}, ErrNotFound
	}
	return m, err
}
