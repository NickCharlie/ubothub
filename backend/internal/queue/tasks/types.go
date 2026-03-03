package tasks

// Task type constants define all async task identifiers.
// Following the convention: <module>:<action>.
const (
	TypeEmailSend         = "email:send"
	TypeAssetProcess      = "asset:process"
	TypeBotHealthCheck    = "bot:health_check"
	TypeMessageDispatch   = "message:dispatch"
	TypeMessageLog        = "message:log"
)

// Queue name constants for priority-based routing.
const (
	QueueCritical = "critical"
	QueueDefault  = "default"
	QueueLow      = "low"
)
