package postgres_repository

import (
	"context"
	"testing"
	"time"
	"uuid"

	postgres_pool "github.com/GinciuuKitakaze/users/internal/adapter/repository/postgres/pool"
)

type fakePool struct {
	row postgres_pool.Row
}

func (p *fakePool) Query(ctx context.Context, sql string, args ...any) (postgres_pool.Rows, error) {
	return nil, nil
}

func (p *fakePool) OpTimeout() time.Duration {
	return time.Second
}

func (p *fakePool) QueryRow(ctx context.Context, sql string, args ...any) postgres_pool.Row {
	return p.row
}

func (p *fakePool) Exec(ctx context.Context, sql string, arguments ...any) (postgres_pool.CommandTag, error) {
	return nil, nil
}

func (p *fakePool) Close() {

}

type fakeRow struct {
	scan func(dest ...any) error
}

func (r fakeRow) Scan(dest ...any) error {
	return r.scan(dest...)
}

func TestRepository_GetUser(t *testing.T) {

	userID := uuid.NewV4()
	fakePoolTest1 := &fakePool{
		row: fakeRow{
			scan: func(dest ...any) error {
				*dest[0].(*uuid.UUID) = userID
				*dest[1].(*int64) = 1
				email := "test.email"
				phone := "test.phone"
				*dest[2].(**string) = &email
				*dest[3].(**string) = &phone

				*dest[4].(*string) = "test"
				*dest[5].(*bool) = false
				*dest[6].(*time.Time) = time.Now()
				*dest[7].(*time.Time) = time.Now()
				return nil
			},
		},
	}

	repository := &Repository{
		pool: fakePoolTest1,
	}

	got, err := repository.GetUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("repository.GetUser(): %v", err)
	}

	if got.ID != userID {
		t.Errorf("repository.GetUser(): got %v, want %v", got.ID, userID)
	}

	if got.Version != 1 {
		t.Errorf("repository.GetUser(): got %v, want %v", got.Version, 1)
	}

	if got.Name != "test" {
		t.Errorf("repository.GetUser(): got %v, want %v", got.Name, "test")
	}

	fakePoolTest2 := &fakePool{
		row: fakeRow{
			scan: func(dest ...any) error {
				return postgres_pool.ErrNoRows
			},
		},
	}

	repository1 := &Repository{
		pool: fakePoolTest2,
	}

	tk, err := repository1.GetUser(context.Background(), userID)
	if err == nil {
		t.Fatalf("repository.GetUser(): got %v, want error", err)
	}
	_ = tk

}
