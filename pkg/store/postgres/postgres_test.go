package postgres

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	"github.com/d7561985/mongo-ab/pkg/changing"
	"github.com/d7561985/pb-ab/internal/config"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	repo, err := New(context.Background(), config.Postgres{
		Addr: "postgresql://postgres@localhost/db"})
	require.NoError(t, err)

	err = repo.Setup(context.Background())
	require.NoError(t, err)

	req := genRequest(1, 100)
	b, err := repo.UpdateTX(context.Background(), req)

	require.NoError(t, err)
	fmt.Println(b)
}

func genRequest(usr uint64, add float64) changing.Transaction {
	tx := changing.Transaction{}
	fuzz.New().Fuzz(&tx)

	tx.Inc = changing.Inc{
		Balance:        add,
		DepositAllSum:  100,
		DepositCount:   1,
		PincoinBalance: 100,
		PincoinsAllSum: 1,
	}

	tx.AccountID = usr
	tx.Currency = 123
	tx.Change = add
	tx.TransactionID = uint64(rand.Int63())
	return tx
}
