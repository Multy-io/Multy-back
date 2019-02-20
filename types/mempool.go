package types

type TransactionFeeRateEstimation struct {
	VerySlow uint64 `bson:"VerySlow" json:"VerySlow"`
	Slow     uint64 `bson:"Slow" json:"Slow"`
	Medium   uint64 `bson:"Medium" json:"Medium"`
	Fast     uint64 `bson:"Fast" json:"Fast"`
	VeryFast uint64 `bson:"VeryFast" json:"VeryFast"`
}
