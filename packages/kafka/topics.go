package kafka

const (
	TopicOrderSubmitted          = "order.submitted"
	TopicOrderCancelled          = "order.cancelled"
	TopicOrderPartiallyFilled    = "order.partially_filled"
	TopicTradeExecuted           = "trade.executed"
	TopicTradeCleared            = "trade.cleared"
	TopicTradeSettled            = "trade.settled"
	TopicRiskRejected            = "risk.rejected"
	TopicCircuitBreakerTriggered = "circuit.breaker.triggered"
	TopicCircuitBreakerLifted    = "circuit.breaker.lifted"
)
