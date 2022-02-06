package postgres

import (
	"time"

	"github.com/d7561985/mongo-ab/pkg/changing"
	"github.com/google/uuid"
)

type Balance struct {
	AccountID     uint64
	Balance       float64
	DepositAllSum float64
	DepositCount  int32

	PincoinBalance float64
	PincoinsAllSum float64
}

type Journal struct {
	ID   uuid.UUID
	ID2  []byte
	Date time.Time

	Balance
	Change        float32
	PincoinChange float32

	TransactionType   string
	TransactionID     int64
	TransactionIDBson []byte
	Type              string
	Project           string
	Currency          int8
	Revert            bool
}

func NewJournal(b Balance, in changing.Transaction) Journal {
	return Journal{
		ID2:               in.Set.ID[:],
		Balance:           b,
		TransactionType:   in.Set.TransactionType,
		TransactionID:     int64(in.Set.TransactionID),
		TransactionIDBson: in.Set.TransactionIDBson[:],
		Date:              in.Set.Date,
		Type:              in.Set.Type,
		Project:           in.Set.Project,
		Currency:          int8(in.Set.Currency),
		PincoinChange:     float32(in.Set.PincoinChange),
		Change:            float32(in.Set.Change),
		Revert:            in.Set.Revert,
	}
}
