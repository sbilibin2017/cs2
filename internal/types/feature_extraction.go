package types

import (
	"sort"
	"time"
)

type FeaturePlayerParams struct {
	BeginAt   time.Time `json:"begin_at"`
	PlayerIDs [10]int64 `json:"player_ids"`
}

type FeaturePlayer struct {
	Value1  float64 `json:"value1"`
	Value2  float64 `json:"value2"`
	Value3  float64 `json:"value3"`
	Value4  float64 `json:"value4"`
	Value5  float64 `json:"value5"`
	Value6  float64 `json:"value6"`
	Value7  float64 `json:"value7"`
	Value8  float64 `json:"value8"`
	Value9  float64 `json:"value9"`
	Value10 float64 `json:"value10"`

	Values1Avg float64 `json:"values1"`
	Values2Avg float64 `json:"values2"`

	Value1Value6Sub  float64 `json:"value1_value6_sub"`
	Value1Value7Sub  float64 `json:"value1_value7_sub"`
	Value1Value8Sub  float64 `json:"value1_value8_sub"`
	Value1Value9Sub  float64 `json:"value1_value9_sub"`
	Value1Value10Sub float64 `json:"value1_value10_sub"`

	Value2Value6Sub  float64 `json:"value2_value6_sub"`
	Value2Value7Sub  float64 `json:"value2_value7_sub"`
	Value2Value8Sub  float64 `json:"value2_value8_sub"`
	Value2Value9Sub  float64 `json:"value2_value9_sub"`
	Value2Value10Sub float64 `json:"value2_value10_sub"`

	Value3Value6Sub  float64 `json:"value3_value6_sub"`
	Value3Value7Sub  float64 `json:"value3_value7_sub"`
	Value3Value8Sub  float64 `json:"value3_value8_sub"`
	Value3Value9Sub  float64 `json:"value3_value9_sub"`
	Value3Value10Sub float64 `json:"value3_value10_sub"`

	Value4Value6Sub  float64 `json:"value4_value6_sub"`
	Value4Value7Sub  float64 `json:"value4_value7_sub"`
	Value4Value8Sub  float64 `json:"value4_value8_sub"`
	Value4Value9Sub  float64 `json:"value4_value9_sub"`
	Value4Value10Sub float64 `json:"value4_value10_sub"`

	Value5Value6Sub  float64 `json:"value5_value6_sub"`
	Value5Value7Sub  float64 `json:"value5_value7_sub"`
	Value5Value8Sub  float64 `json:"value5_value8_sub"`
	Value5Value9Sub  float64 `json:"value5_value9_sub"`
	Value5Value10Sub float64 `json:"value5_value10_sub"`
}

func NewFeaturePlayer(values []float64) *FeaturePlayer {
	padded := make([]float64, 10)
	copy(padded, values)

	sort.Float64s(padded[:5])
	sort.Float64s(padded[5:])

	var sum1, sum2 float64
	for i := 0; i < 5; i++ {
		sum1 += padded[i]
		sum2 += padded[i+5]
	}
	avg1 := sum1 / 5
	avg2 := sum2 / 5

	return &FeaturePlayer{
		Value1:  padded[0],
		Value2:  padded[1],
		Value3:  padded[2],
		Value4:  padded[3],
		Value5:  padded[4],
		Value6:  padded[5],
		Value7:  padded[6],
		Value8:  padded[7],
		Value9:  padded[8],
		Value10: padded[9],

		Values1Avg: avg1,
		Values2Avg: avg2,

		Value1Value6Sub:  padded[0] - padded[5],
		Value1Value7Sub:  padded[0] - padded[6],
		Value1Value8Sub:  padded[0] - padded[7],
		Value1Value9Sub:  padded[0] - padded[8],
		Value1Value10Sub: padded[0] - padded[9],

		Value2Value6Sub:  padded[1] - padded[5],
		Value2Value7Sub:  padded[1] - padded[6],
		Value2Value8Sub:  padded[1] - padded[7],
		Value2Value9Sub:  padded[1] - padded[8],
		Value2Value10Sub: padded[1] - padded[9],

		Value3Value6Sub:  padded[2] - padded[5],
		Value3Value7Sub:  padded[2] - padded[6],
		Value3Value8Sub:  padded[2] - padded[7],
		Value3Value9Sub:  padded[2] - padded[8],
		Value3Value10Sub: padded[2] - padded[9],

		Value4Value6Sub:  padded[3] - padded[5],
		Value4Value7Sub:  padded[3] - padded[6],
		Value4Value8Sub:  padded[3] - padded[7],
		Value4Value9Sub:  padded[3] - padded[8],
		Value4Value10Sub: padded[3] - padded[9],

		Value5Value6Sub:  padded[4] - padded[5],
		Value5Value7Sub:  padded[4] - padded[6],
		Value5Value8Sub:  padded[4] - padded[7],
		Value5Value9Sub:  padded[4] - padded[8],
		Value5Value10Sub: padded[4] - padded[9],
	}
}
