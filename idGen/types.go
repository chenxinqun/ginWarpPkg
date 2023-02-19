package idGen

type IDGenerator interface {
	NextID() (int64, error)
	GetDeviceID(sid int64) (datacenterid, workerid int64)
	GetTimestamp(sid int64) (timestamp int64)
	GetGenTimestamp(sid int64) (timestamp int64)
	GetGenTime(sid int64, layout string) (t string)
	GetTimestampStatus() (state float64)
}

var defaultIDGenerator IDGenerator

func Default() IDGenerator {
	return defaultIDGenerator
}
