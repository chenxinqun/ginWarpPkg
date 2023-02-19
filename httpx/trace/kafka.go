package trace

type Kafka struct {
	Timestamp   string  `json:"timestamp"`    // 时间，格式：2006-01-02 15:04:05
	Handle      string  `json:"handle"`       // 操作，CURD
	Topic       string  `json:"topic"`        // 操作的topic
	Offset      int64   `json:"offset"`       // offset
	TTL         float64 `json:"ttl"`          // 超时时长(单位分)
	CostSeconds float64 `json:"cost_seconds"` // 执行时间(单位秒)
}
