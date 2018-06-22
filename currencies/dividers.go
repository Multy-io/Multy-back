package currencies

const (
	Satoshi = int64(100000000)
	Wei     = int64(1000000000000000000)
)

var Dividers = map[int]int64{
	Bitcoin: Satoshi,
	Ether:   Wei,
}
