package log

type Config struct {
	Segment struct {
		MaxStoreByte  uint64
		MaxIndexBytes uint64
		InitialOffset uint64
	}
}
