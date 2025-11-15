package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shirr9/pr-reviewer-service/internal/app/config"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(ctx context.Context, cfg config.Config) (*Storage, error) {
	// dbdriver://username:password@host:port/dbname?param1=true&param2=false
	dsn := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.PostgresDb.Username, cfg.PostgresDb.Password, cfg.PostgresDb.Host,
		cfg.PostgresDb.Port, cfg.PostgresDb.DbName, cfg.PostgresDb.SSlMode)
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection Pool: %w", err)
	}
	if e := pool.Ping(ctx); e != nil {
		return nil, fmt.Errorf("failed to ping database: %w", e)
	}
	return &Storage{pool: pool}, nil
}

func (s *Storage) NewUnitOfWork() *UnitOfWork {
	return &UnitOfWork{pool: s.pool}
}

func (s *Storage) NewPullRequestRepository() *PullRequestRepository {
	return &PullRequestRepository{pool: s.pool}
}

func (s *Storage) NewReviewerRepository() *ReviewerRepository {
	return &ReviewerRepository{pool: s.pool}
}

func (s *Storage) NewTeamRepository() *TeamRepository {
	return &TeamRepository{pool: s.pool}
}

func (s *Storage) NewUserRepository() *UserRepository {
	return &UserRepository{pool: s.pool}
}

func (s *Storage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}
