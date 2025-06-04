package types

type TrainTestSplit struct {
	TrainIDs []int `json:"train_ids"`
	TestIDs  []int `json:"test_ids"`
}
