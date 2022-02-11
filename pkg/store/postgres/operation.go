package postgres

// OpType define current operation
// Warning! Don't forget update enum in database when you add/modify any of that
type OpType string

const (
	None           OpType = "None"
	Deposit        OpType = "Add Deposit"
	Bet            OpType = "Write bet"
	FreebetWin     OpType = "FreebetWin"
	Withdraw       OpType = "Withdraw"
	LotteryWin     OpType = "LotteryWin"
	WelcomeDeposit OpType = "Welcome deposit"
	Revert         OpType = "Revert"
)

func NewOperationType(in string) OpType {
	switch v := OpType(in); v {
	case Deposit, Bet, FreebetWin, Withdraw, LotteryWin, WelcomeDeposit, Revert:
		return v
	default:
		return None
	}
}
