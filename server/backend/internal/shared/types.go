package shared

import (
	"crypto/rand"
	"math"
	"math/big"
)

type WsNewTaskRequest struct {
	Cmd        string `json:"cmd"`
	Args       []Arg  `json:"args"`
	ClientID   int32  `json:"client_id"`
	CampaignID string `json:"campaign_id"`
}

type Arg struct {
	Type  string `json:"type"`
	Value any    `json:"value"`
}

func GenRandInt32() (int32, error) {
	const min int64 = 1_500_000_000
	const max int64 = math.MaxInt32
	n, err := rand.Int(rand.Reader, big.NewInt(max-min+1))
	if err != nil {
		return 0, err
	}
	return int32(min + n.Int64()), nil
}
