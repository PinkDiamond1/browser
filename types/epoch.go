package types

type Epochs struct {
	Data []*Epoch `json:"data"`
}

type Epoch struct {
	Start uint64 `json:"start"`
	Epoch uint64 `json:"epoch"`
}

// ArrayCandidateInfoForBrowser dpos state
type ArrayCandidateInfoForBrowser struct {
	Data     []*CandidateInfoForBrowser `json:"data"`
	Bad      []uint64                   `json:"bad"`
	Using    []uint64                   `json:"using"`
	TakeOver bool                       `json:"takeOver"`
	Dpos     bool                       `json:"dpos"`
}

// CandidateInfoForBrowser dpos state
type CandidateInfoForBrowser struct {
	Candidate        string `json:"candidate"`
	Holder           string `json:"holder"`
	Quantity         string `json:"quantity"`
	TotalQuantity    string `json:"totalQuantity"`
	Counter          uint64 `json:"shouldCounter"`
	ActualCounter    uint64 `json:"actualCounter"`
	NowCounter       uint64 `json:"nowShouldCounter"`
	NowActualCounter uint64 `json:"nowActualCounter"`
	// Status           uint64 `json:"status"` //0:die 1:activate 2:spare
	Epoch    uint64
	Activate uint64
	Die      uint64
	Spare    uint64
	Rank     uint64
	Replace  uint64
	Vote     int
}

type ChangeIng struct {
	Epoch uint64
	Info  *ArrayCandidateInfoForBrowser
}

// EpochReward .
type EpochReward struct {
	Rewards     []*Reward `json:"rewards"` //周期奖励
	GiveOutTime int64     `json:"time"`    //奖励时间
	Amount      string    `json:"amount"`  //奖励FT数量
	LockRatio   int64     `json:"ratio"`   //锁仓占比
	Index       int64     `json:"index"`   //奖励索引
}

// LastReward .
type LastReward struct {
	Epoch       uint64    `json:"epoch"`   //周期
	Rewards     []*Reward `json:"rewards"` //周期奖励
	GiveOutTime int64     `json:"time"`    //奖励时间
	Amount      string    `json:"amount"`  //奖励FT数量
	LockRatio   int64     `json:"ratio"`   //锁仓占比
	Index       int64     `json:"index"`   //奖励索引
}

// Reward .
type Reward struct {
	Candidate     string `json:"ca"` //candidate生产者
	OriginalRank  uint64 `json:"or"` //originalRank初始排名
	Counter       uint64 `json:"sc"` //shouldCounter
	ActualCounter uint64 `json:"ac"` //actualCounter
	RewardRatio   string `json:"rr"` //rewardRatio奖励占比
	AccoundReward string `json:"ar"` //accoundReward记账奖励
	VoteReward    string `json:"vr"` //voteReward投票奖励
	TotalQuantity string `json:"tq"` //totalQuantity得票数
	ReturnRate    string `json:"re"` //returnRate投票回报率
	Weight        uint64 `json:"we"` //weight权重
}
