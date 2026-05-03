package fix

const (
	TagBeginString  = 8
	TagBodyLength   = 9
	TagMsgType      = 35
	TagSenderCompID = 49
	TagTargetCompID = 56
	TagMsgSeqNum    = 34
	TagSendingTime  = 52
	TagClOrdID      = 11
	TagSymbol       = 55
	TagSide         = 54
	TagOrderQty     = 38
	TagOrdType      = 40
	TagPrice        = 44
	TagTimeInForce  = 59
	TagCheckSum     = 10
	TagAPIKey       = 553
)

const (
	MsgTypeNewOrder        = "D"
	MsgTypeCancelOrder     = "F"
	MsgTypeExecutionReport = "8"
	MsgTypeReject          = "3"
)

const (
	SideBuy  = "1"
	SideSell = "2"
)

const (
	OrdTypeMarket = "1"
	OrdTypeLimit  = "2"
	OrdTypeStop   = "3"
)

const (
	TimeInForceGTC = "1"
	TimeInForceIOC = "3"
)

const (
	ExecTypeNew         = "0"
	ExecTypeFilled      = "2"
	ExecTypePartialFill = "1"
	ExecTypeRejected    = "8"
	ExecTypeCancelled   = "4"
)

const (
	OrdStatusNew         = "0"
	OrdStatusFilled      = "2"
	OrdStatusPartialFill = "1"
	OrdStatusRejected    = "8"
	OrdStatusCancelled   = "4"
)
