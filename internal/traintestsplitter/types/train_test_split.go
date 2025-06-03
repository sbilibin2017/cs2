package types

type TrainTestSplit struct {
	Hash         string  `json:"split_hash"`
	TrainGameIDs []int32 `json:"train_game_ids"`
	TestGameIDs  []int32 `json:"test_game_ids"`
}
