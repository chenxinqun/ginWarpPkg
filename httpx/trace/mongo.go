package trace

type Mongo struct {
	Timestamp   string  `json:"timestamp"`      // 时间，格式：2006-01-02 15:04:05
	Handle      string  `json:"handle"`         // 操作，CURD
	Collection  string  `json:"collection"`     // 操作的MongoDB表
	Count       int64   `json:"count_affected"` // 影响文档数
	TTL         float64 `json:"ttl"`            // 超时时长(单位分)
	CostSeconds float64 `json:"cost_seconds"`   // 执行时间(单位秒)
}
