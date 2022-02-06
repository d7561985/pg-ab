package postgres

import (
	"context"

	"github.com/d7561985/mongo-ab/pkg/changing"
	"github.com/d7561985/pb-ab/internal/config"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/pkg/errors"
)

const maxCon = 100

type Repo struct {
	cfg config.Postgres

	pool *pgxpool.Pool
}

func New(ctx context.Context, cfg config.Postgres) (*Repo, error) {
	c, err := pgxpool.ParseConfig(cfg.Addr)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	c.MaxConns = maxCon

	dbpool, err := pgxpool.ConnectConfig(ctx, c)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if err = dbpool.Ping(ctx); err != nil {
		return nil, errors.WithStack(err)
	}

	return &Repo{
		cfg:  cfg,
		pool: dbpool,
	}, nil
}

func (s *Repo) UpdateTX(ctx context.Context, in changing.Transaction) (_ interface{}, err error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer func() {
		if err == nil {
			err = errors.WithStack(tx.Commit(ctx))
		} else {
			_ = tx.Rollback(ctx)
		}
	}()

	res := tx.QueryRow(ctx, `INSERT INTO balance("accountId", "balance", "depositAllSum",
                    "depositCount", "pincoinBalance", "pincoinAllSum") VALUES ($1,$2,$3,$4,$5,$6) 
		ON CONFLICT ON CONSTRAINT balance_pkey DO UPDATE SET 
			balance = balance.balance + $2,
			"depositAllSum" = balance."depositAllSum" + $3,
            "depositCount" = balance."depositCount" + $4,
			"pincoinBalance" = balance."pincoinBalance" + $5,
			"pincoinAllSum" = balance."pincoinAllSum" + $6
			WHERE balance."accountId" = $1 
			RETURNING "balance","depositAllSum", "depositCount", "pincoinBalance", "pincoinAllSum"`,
		in.AccountID, in.Balance, in.DepositAllSum, in.DepositCount, in.PincoinBalance, in.PincoinsAllSum)

	b := Balance{AccountID: in.AccountID}
	if err = res.Scan(&b.Balance, &b.DepositAllSum, &b.DepositCount, &b.PincoinBalance, &b.PincoinsAllSum); err != nil {
		return nil, errors.WithStack(err)
	}

	j := NewJournal(b, in)
	sq := `INSERT INTO journal("id","id2","accountId","balance","change","currency","created_at","depositAllSum","depositCount",
                "pincoinBalance","pincoinAllSum","pincoinChange","project","revert","transactionId",
                "transactionBson", "transactionType"
            ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15, $16, $17)`
	_, err = tx.Exec(ctx, sq,
		uuid.New(), j.ID2, j.AccountID, j.Balance.Balance, j.Change, j.Currency, j.Date, j.DepositAllSum, j.DepositCount,
		j.PincoinBalance, j.PincoinsAllSum, j.PincoinChange, j.Project, j.Revert, j.TransactionID,
		j.TransactionIDBson, j.TransactionType,
	)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return b, nil
}

func (s *Repo) Insert(ctx context.Context, j Journal) error {
	sq := `INSERT INTO journal("id","id2","accountId","balance","change","currency","created_at","depositAllSum","depositCount",
                "pincoinBalance","pincoinAllSum","pincoinChange","project","revert","transactionId",
                "transactionBson", "transactionType"
            ) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15, $16, $17)`
	_, err := s.pool.Exec(ctx, sq,
		uuid.New(), j.ID2, j.AccountID, j.Balance.Balance, j.Change, j.Currency, j.Date, j.DepositAllSum, j.DepositCount,
		j.PincoinBalance, j.PincoinsAllSum, j.PincoinChange, j.Project, j.Revert, j.TransactionID,
		j.TransactionIDBson, j.TransactionType,
	)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil

}
