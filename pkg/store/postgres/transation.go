package postgres

import (
	"encoding/binary"
	"time"
	"unsafe"

	"github.com/d7561985/mongo-ab/pkg/changing"
)

const sizeUint64 = int(unsafe.Sizeof(uint64(0)))

type Balance struct {
	AccountID     uint64
	Balance       float64
	DepositAllSum float64
	DepositCount  int32

	PincoinBalance float64
	PincoinsAllSum float64
}

type Journal struct {
	ID []byte

	TransactionID []byte

	AccountID uint64

	CreatedAt time.Time

	Balance float64
	Change  float32

	PincoinBalance float64
	PincoinChange  float32

	Type OpType

	Project  Project
	Currency int8

	Revert bool
}

func NewJournal(b Balance, in changing.Transaction) Journal {
	trn, err := GetTransactionID(in)
	if err != nil {
		trn = []byte{}
	}

	return Journal{
		ID:            in.Set.ID[:],
		TransactionID: trn,
		AccountID:     b.AccountID,

		CreatedAt: time.Now(),

		Balance:        b.Balance,
		Change:         float32(in.Change),
		PincoinBalance: b.PincoinBalance,
		PincoinChange:  float32(in.PincoinChange),

		Type:     NewOperationType(in.Type),
		Project:  NewProject(in.Project),
		Currency: int8(in.Currency),
		Revert:   false,
	}
}

func GetTransactionID(r changing.Transaction) ([]byte, error) {
	if r.TransactionID > 0 {
		tID := make([]byte, sizeUint64)
		binary.LittleEndian.PutUint64(tID, r.TransactionID)

		return tID, nil
	}

	return r.TransactionIDBson[:], nil
}
