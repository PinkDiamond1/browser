package mem

import "fmt"

// import (
// 	"fmt"
// 	"math/big"
// 	"strconv"
// 	"sync"
// 	"time"

// 	"github.com/browser/client"
// 	"github.com/browser/common"
// 	. "github.com/browser/log"
// 	"github.com/browser/types"
// 	"go.uber.org/zap"
// )

// type LastReward struct {
// 	Epoch uint64
// 	Index int
// }

// type Dpos struct {
// 	AllEpochs     *types.Epochs
// 	UpdateTime    int64
// 	AllEpochsLock sync.RWMutex

// 	DBAccounds     map[uint64]*types.ArrayCandidateInfoForBrowser
// 	DBAccoundsLock sync.RWMutex

// 	Accounding     *types.ChangeIng
// 	AccoundingLock sync.RWMutex

// 	DBVotes     map[uint64]*types.ArrayCandidateInfoForBrowser
// 	DBVotesLock sync.RWMutex

// 	Voteing     *types.ChangeIng
// 	VoteingLock sync.RWMutex

// 	dbActivate map[string]uint64
// 	dbSpare    map[string]uint64
// 	dbDie      map[string]uint64

// 	accoundingActivate map[string]uint64
// 	accoundingSpare    map[string]uint64
// 	accoundingDie      map[string]uint64

// 	EpochReward     map[uint64][]*types.EpochReward
// 	MaxIndex        int64
// 	LastRewards     []*LastReward
// 	CompletedReward map[int64]uint64
// 	EpochRewardLock sync.RWMutex
// }

// var (
// 	dposSyncDuration = 3 * time.Second
// 	spare            = 7
// 	activate         = 21
// )

// var MemDpos *Dpos

// var accountMap = make(map[uint64]string, 0)

func DposStart() {
	fmt.Println("**********")
}

// 	MemDpos = &Dpos{}
// 	MemDpos.DBAccounds = make(map[uint64]*types.ArrayCandidateInfoForBrowser, 0)
// 	MemDpos.Accounding = &types.ChangeIng{}
// 	MemDpos.DBVotes = make(map[uint64]*types.ArrayCandidateInfoForBrowser, 0)
// 	MemDpos.Voteing = &types.ChangeIng{}
// 	MemDpos.dbActivate = make(map[string]uint64, 0)
// 	MemDpos.dbSpare = make(map[string]uint64, 0)
// 	MemDpos.dbDie = make(map[string]uint64, 0)

// 	MemDpos.CompletedReward = make(map[int64]uint64, 0)
// 	MemDpos.EpochReward = make(map[uint64][]*types.EpochReward, 0)
// 	load()

// 	now := time.Now()
// 	nextRwward := now.Add(time.Hour)
// 	// nextRwward = time.Date(nextRwward.Year(), nextRwward.Month(), nextRwward.Day(), 0, 0, 0, 0, nextRwward.Location())
// 	go reward()

// 	time.Sleep(dposSyncDuration)
// 	for {
// 		epochs, _ := client.GetBrowserAllEpoch()
// 		for i, e := range epochs.Data {
// 			if i == 0 {
// 				MemDpos.AccoundingLock.RLock()
// 				old := MemDpos.Accounding.Epoch
// 				MemDpos.UpdateTime = time.Now().Unix()
// 				MemDpos.AccoundingLock.RUnlock()
// 				if old < e.Epoch {
// 					MemDpos.AllEpochsLock.Lock()
// 					MemDpos.AllEpochs = epochs
// 					MemDpos.AllEpochsLock.Unlock()
// 					tryInsertAccoundToMysql(old)
// 					tryInsertVoteToMysql(old)
// 				}
// 				setAccounding(e.Epoch)
// 				setVoteing(e.Epoch)
// 			} else {
// 				MemDpos.DBAccoundsLock.RLock()
// 				_, ok := MemDpos.DBAccounds[e.Epoch]
// 				MemDpos.DBAccoundsLock.RUnlock()
// 				if !ok {
// 					panic(e.Epoch)
// 				}
// 			}
// 		}

// 		if time.Now().Unix() > nextRwward.Unix() {
// 			nextRwward = nextRwward.Add(time.Hour)
// 			go reward()
// 		}
// 		time.Sleep(dposSyncDuration)
// 	}
// }

// func load() {
// 	epochs, _ := client.GetBrowserAllEpoch()
// 	MemDpos.AllEpochsLock.Lock()
// 	MemDpos.AllEpochs = epochs
// 	MemDpos.AllEpochsLock.Unlock()

// 	count := len(epochs.Data)
// 	for i := count - 1; i >= 0; i-- {
// 		e := epochs.Data[i]
// 		if i == 0 {
// 			ZapLog.Info("Load SetAccounding", zap.Uint64("epoch", e.Epoch))
// 			setAccounding(e.Epoch)
// 			setVoteing(e.Epoch)
// 		} else {
// 			ZapLog.Info("Load DB", zap.Uint64("epoch", e.Epoch))
// 			tryInsertAccoundToMysql(e.Epoch)
// 			tryInsertVoteToMysql(e.Epoch)
// 		}
// 	}
// }

// func reward() {
// 	epochs := make([]uint64, 0)

// 	MemDpos.AllEpochsLock.RLock()
// 	for _, e := range MemDpos.AllEpochs.Data {
// 		epochs = append(epochs, e.Epoch)
// 	}
// 	MemDpos.AllEpochsLock.RUnlock()

// 	for _, epoch := range epochs {
// 		cycle, _ := getCycleInfo(int64(epoch))
// 		if len(cycle.Indexs) == 0 {
// 			continue
// 		}

// 		for _, index := range cycle.Indexs {
// 			_, ok := MemDpos.CompletedReward[index.Int64()]
// 			if ok {
// 				continue
// 			}
// 			data := &types.EpochReward{}
// 			tmpRewards := make([]*types.Reward, 0)
// 			MemDpos.DBAccoundsLock.RLock()
// 			candidateInfos, ok := MemDpos.DBAccounds[epoch]
// 			if ok {
// 				for _, e := range candidateInfos.Data {
// 					tmp := &types.Reward{}
// 					tmp.Candidate = e.Candidate
// 					tmp.TotalQuantity = e.TotalQuantity
// 					tmp.Counter = e.NowCounter - e.Counter
// 					tmp.ActualCounter = e.NowActualCounter - e.ActualCounter
// 					tmpRewards = append(tmpRewards, tmp)
// 				}
// 			} else {
// 				ZapLog.Panic(fmt.Sprintf("epoch：%d", epoch))
// 			}
// 			MemDpos.DBAccoundsLock.RUnlock()

// 			rewardInfo, _ := getRewardInfo(index)
// 			data.Index = index.Int64()
// 			data.LockRatio = rewardInfo.LockRatio.Int64()
// 			data.GiveOutTime = rewardInfo.Time.Int64() / 1000000000
// 			data.Amount = rewardInfo.Amount.String()
// 			if len(tmpRewards) != len(rewardInfo.SingleValue) {
// 				ZapLog.Panic(fmt.Sprintf("Rewards len:%d SingleValue len:%d", len(tmpRewards), len(rewardInfo.SingleValue)))
// 			}
// 			for i, signalValue := range rewardInfo.SingleValue {
// 				tmpRewards[i].ReturnRate = big.NewInt(0).Mul(signalValue, big.NewInt(100)).String()
// 			}

// 			producersName := make([]string, 0)
// 			for _, producer := range cycle.Producers {
// 				name, ok := accountMap[producer.Uint64()]
// 				if !ok {
// 					name, _ = client.GetAccountByID(producer.Uint64())
// 					accountMap[producer.Uint64()] = name
// 				}
// 				producersName = append(producersName, name)
// 			}

// 			data.Rewards = make([]*types.Reward, 0)
// 			for _, name := range producersName {
// 				for j, r := range tmpRewards {
// 					if r.Candidate == name {
// 						r.OriginalRank = uint64(j) + 1
// 						data.Rewards = append(data.Rewards, r)
// 						break
// 					}
// 				}
// 			}

// 			if len(tmpRewards) != len(data.Rewards) {
// 				ZapLog.Panic(fmt.Sprintf("Rewards len:%d data.Rewards len:%d", len(tmpRewards), len(data.Rewards)))
// 			}

// 			weightsSum := big.NewInt(0)
// 			for _, weight := range rewardInfo.Weights {
// 				weightsSum = weightsSum.Add(weightsSum, weight)
// 			}

// 			for i, weight := range rewardInfo.Weights {
// 				producerReward := big.NewInt(0).Mul(rewardInfo.Amount, weight)
// 				producerReward = producerReward.Div(producerReward, weightsSum)
// 				producerRewardCopy, _ := big.NewInt(0).SetString(producerReward.String(), 10)
// 				producerReward = producerReward.Div(producerReward, types.Big100)
// 				producerReward = producerReward.Mul(producerReward, types.Big80)

// 				producerRewardCopy = producerRewardCopy.Div(producerRewardCopy, types.Big100)
// 				producerRewardCopy = producerRewardCopy.Mul(producerRewardCopy, types.Big20)

// 				data.Rewards[i].AccoundReward = producerReward.String()
// 				data.Rewards[i].VoteReward = producerRewardCopy.String()
// 				data.Rewards[i].Weight = weight.Uint64()

// 				rewardRatio := float64(weight.Int64()) / float64(weightsSum.Int64())
// 				data.Rewards[i].RewardRatio = strconv.FormatFloat(rewardRatio, 'f', -1, 64)
// 			}

// 			MemDpos.CompletedReward[index.Int64()] = 0

// 			MemDpos.EpochRewardLock.Lock()
// 			_, ok = MemDpos.EpochReward[epoch]
// 			if !ok {
// 				MemDpos.EpochReward[epoch] = make([]*types.EpochReward, 0)
// 			}
// 			MemDpos.EpochReward[epoch] = append(MemDpos.EpochReward[epoch], data)
// 			if MemDpos.MaxIndex < index.Int64() {
// 				MemDpos.MaxIndex = index.Int64()
// 			}
// 			MemDpos.EpochRewardLock.Unlock()
// 		}
// 	}

// 	MemDpos.EpochRewardLock.Lock()
// 	MemDpos.LastRewards = make([]*LastReward, 0)
// 	max := MemDpos.MaxIndex
// 	for {
// 		ok := false
// 		for epoch, epochRewards := range MemDpos.EpochReward {
// 			if len(MemDpos.LastRewards) > 2 {
// 				break
// 			}
// 			for i, reward := range epochRewards {
// 				if reward.Index == max {
// 					tmp := &LastReward{}
// 					tmp.Epoch = epoch
// 					tmp.Index = i
// 					MemDpos.LastRewards = append(MemDpos.LastRewards, tmp)
// 					max = max - 1
// 					ok = true
// 				}
// 			}
// 		}
// 		if !ok {
// 			break
// 		}
// 	}

// 	MemDpos.EpochRewardLock.Unlock()
// }

// func setVoteing(epoch uint64) {
// 	candidates, _ := client.GetBrowserVote(epoch)
// 	var declims uint64 = 1000000000000000000
// 	var min uint64 = 1000000
// 	minQuantity := big.NewInt(0).Mul(big.NewInt(0).SetUint64(min), big.NewInt(0).SetUint64(declims))
// 	for _, candidate := range candidates.Data {
// 		holder, _ := big.NewInt(0).SetString(candidate.Holder, 10)
// 		if holder.Cmp(minQuantity) != -1 {
// 			candidate.Vote = 1
// 		}
// 		candidate.Epoch = epoch
// 		candidate.Holder = candidate.Holder
// 		candidate.Quantity = candidate.Quantity
// 		candidate.Activate = MemDpos.dbActivate[candidate.Candidate] + MemDpos.accoundingActivate[candidate.Candidate]
// 		candidate.Die = MemDpos.dbDie[candidate.Candidate] + MemDpos.accoundingDie[candidate.Candidate]
// 		candidate.Spare = MemDpos.dbSpare[candidate.Candidate] + MemDpos.accoundingSpare[candidate.Candidate]
// 	}

// 	MemDpos.VoteingLock.Lock()
// 	MemDpos.Voteing.Epoch = epoch
// 	MemDpos.Voteing.Info = candidates
// 	MemDpos.VoteingLock.Unlock()
// }

// func tryInsertVoteToMysql(epoch uint64) {
// 	candidates, _ := client.GetBrowserVote(epoch)

// 	var declims uint64 = 1000000000000000000
// 	var min uint64 = 1000000
// 	minQuantity := big.NewInt(0).Mul(big.NewInt(0).SetUint64(min), big.NewInt(0).SetUint64(declims))
// 	for _, candidate := range candidates.Data {
// 		candidate.Epoch = epoch
// 		holder, _ := big.NewInt(0).SetString(candidate.Holder, 10)
// 		if holder.Cmp(minQuantity) != -1 {
// 			candidate.Vote = 1
// 		}
// 		candidate.Holder = candidate.Holder
// 		candidate.Quantity = candidate.Quantity
// 		candidate.Activate = MemDpos.dbActivate[candidate.Candidate]
// 		candidate.Die = MemDpos.dbDie[candidate.Candidate]
// 		candidate.Spare = MemDpos.dbSpare[candidate.Candidate]
// 	}
// 	MemDpos.DBVotesLock.Lock()
// 	MemDpos.DBVotes[epoch] = candidates
// 	MemDpos.DBVotesLock.Unlock()
// }

// func setAccounding(epoch uint64) {
// 	candidates, _ := client.GetBrowserEpochRecord(epoch)

// 	MemDpos.accoundingActivate = make(map[string]uint64, 0)
// 	MemDpos.accoundingSpare = make(map[string]uint64, 0)
// 	MemDpos.accoundingDie = make(map[string]uint64, 0)

// 	// candidates.TakeOver, _ = client.GetTakeOverNewest()
// 	for j, candidate := range candidates.Data {
// 		candidate.Epoch = epoch
// 		candidate.Holder = candidate.Holder
// 		candidate.Quantity = candidate.Quantity
// 		candidate.Replace = 10000
// 		if j < activate {
// 			candidate.Activate = 1
// 			candidate.Spare = 0
// 			MemDpos.accoundingActivate[candidate.Candidate]++
// 		} else {
// 			candidate.Activate = 0
// 			candidate.Spare = 1
// 			MemDpos.accoundingSpare[candidate.Candidate]++
// 		}
// 	}

// 	var tmpArray = make([]uint64, len(candidates.Data))
// 	for i := 0; i < len(candidates.Data); i++ {
// 		tmpArray[i] = uint64(i)
// 	}

// 	for i := 0; i < len(candidates.Bad); i++ {
// 		j := i + activate
// 		if i < spare && len(candidates.Data) > j {
// 			candidates.Data[j].Activate = 1
// 			MemDpos.accoundingActivate[candidates.Data[j].Candidate]++

// 			candidates.Data[j].Replace = tmpArray[candidates.Bad[i]] + 1

// 			dieIndex := tmpArray[candidates.Bad[i]]
// 			candidates.Data[dieIndex].Die = 1
// 			MemDpos.accoundingDie[candidates.Data[dieIndex].Candidate]++
// 			tmpArray[candidates.Bad[i]], tmpArray[j] = tmpArray[j], tmpArray[candidates.Bad[i]]
// 		} else {
// 			dieIndex := tmpArray[candidates.Bad[i]]
// 			candidates.Data[dieIndex].Die = 1
// 			MemDpos.accoundingDie[candidates.Data[dieIndex].Candidate]++
// 		}
// 	}

// 	usingMap := make(map[uint64]int, 0)

// 	for _, index := range candidates.Using {
// 		usingMap[index] = 1
// 	}

// 	for i, c := range candidates.Data {
// 		_, ok := usingMap[uint64(i)]
// 		if !ok {
// 			if c.Die != 1 && c.Activate == 1 {
// 				c.Die = 1
// 				MemDpos.accoundingDie[c.Candidate]++
// 			}
// 		}
// 	}

// 	rank(candidates)

// 	MemDpos.AccoundingLock.Lock()
// 	MemDpos.Accounding.Epoch = epoch
// 	MemDpos.Accounding.Info = candidates
// 	MemDpos.AccoundingLock.Unlock()
// }

// func tryInsertAccoundToMysql(epoch uint64) {
// 	candidates, _ := client.GetBrowserEpochRecord(epoch)

// 	// candidates.TakeOver, _ = client.GetTakeOver(epoch)
// 	for j, candidate := range candidates.Data {
// 		candidate.Epoch = epoch
// 		candidate.Replace = 10000
// 		if j < activate {
// 			candidate.Activate = 1
// 			candidate.Spare = 0
// 			MemDpos.dbActivate[candidate.Candidate]++
// 		} else {
// 			candidate.Activate = 0
// 			candidate.Spare = 1
// 			MemDpos.dbSpare[candidate.Candidate]++
// 		}

// 		candidate.Holder = candidate.Holder
// 		candidate.Quantity = candidate.Quantity
// 	}

// 	var tmpArray = make([]uint64, len(candidates.Data))
// 	for i := 0; i < len(candidates.Data); i++ {
// 		tmpArray[i] = uint64(i)
// 	}

// 	for i := 0; i < len(candidates.Bad); i++ {
// 		j := i + activate
// 		if i < spare && len(candidates.Data) > j {
// 			candidates.Data[j].Activate = 1
// 			MemDpos.dbActivate[candidates.Data[j].Candidate]++

// 			candidates.Data[j].Replace = tmpArray[candidates.Bad[i]] + 1

// 			dieIndex := tmpArray[candidates.Bad[i]]
// 			candidates.Data[dieIndex].Die = 1
// 			MemDpos.dbDie[candidates.Data[dieIndex].Candidate]++
// 			tmpArray[candidates.Bad[i]], tmpArray[j] = tmpArray[j], tmpArray[candidates.Bad[i]]
// 		} else {
// 			dieIndex := tmpArray[candidates.Bad[i]]
// 			candidates.Data[dieIndex].Die = 1
// 			MemDpos.dbDie[candidates.Data[dieIndex].Candidate]++
// 		}
// 	}

// 	usingMap := make(map[uint64]int, 0)

// 	for _, index := range candidates.Using {
// 		usingMap[index] = 1
// 	}

// 	for i, c := range candidates.Data {
// 		_, ok := usingMap[uint64(i)]
// 		if !ok {
// 			if c.Die != 1 && c.Activate == 1 {
// 				c.Die = 1
// 				MemDpos.dbDie[c.Candidate]++
// 			}
// 		}
// 	}

// 	rank(candidates)

// 	MemDpos.DBAccoundsLock.Lock()
// 	MemDpos.DBAccounds[epoch] = candidates
// 	MemDpos.DBAccoundsLock.Unlock()
// }

// func rank(candidates *types.ArrayCandidateInfoForBrowser) {
// 	rankActivate := make([]common.Uint64Sort, 0)
// 	rankSpare := make([]common.Uint64Sort, 0)
// 	rankDie := make([]common.Uint64Sort, 0)

// 	candidateIndex := make(map[string]int)
// 	for i, c := range candidates.Data {
// 		counter := c.NowCounter - c.Counter
// 		actualCounter := c.NowActualCounter - c.ActualCounter
// 		var u common.Uint64Sort

// 		if counter == 0 {
// 			u = common.Uint64Sort{c.Candidate, 0}
// 		} else {
// 			if float64(actualCounter)/float64(counter) > float64(0.95) {
// 				u = common.Uint64Sort{c.Candidate, counter * 1000000 / counter}
// 			} else {
// 				u = common.Uint64Sort{c.Candidate, actualCounter * 1000000 / counter}
// 			}
// 		}

// 		if c.Die == 1 {
// 			rankDie = append(rankDie, u)
// 		} else if c.Activate == 1 {
// 			rankActivate = append(rankActivate, u)
// 		} else if c.Spare == 1 {
// 			rankSpare = append(rankSpare, u)
// 		}
// 		candidateIndex[c.Candidate] = i
// 	}
// 	common.Uint64SorterProcess(rankDie)
// 	common.Uint64SorterProcess(rankActivate)
// 	common.Uint64SorterProcess(rankSpare)

// 	var rank uint64 = 1
// 	var socre uint64
// 	var first = true
// 	equalList := make([]common.Uint64Sort, 0)
// 	for _, e := range rankActivate {
// 		if first {
// 			socre = e.Value
// 			first = false
// 			index := candidateIndex[e.Name]
// 			equalList = append(equalList, common.Uint64Sort{e.Name, uint64(index)})
// 		} else {
// 			if socre == e.Value {
// 			} else {
// 				common.Uint64SorterProcess(equalList)
// 				for i := len(equalList) - 1; i >= 0; i-- {
// 					tmp := equalList[i]
// 					index := candidateIndex[tmp.Name]
// 					candidates.Data[index].Rank = rank
// 					rank++
// 				}
// 				equalList = make([]common.Uint64Sort, 0)
// 				socre = e.Value
// 			}
// 			index := candidateIndex[e.Name]
// 			equalList = append(equalList, common.Uint64Sort{e.Name, uint64(index)})
// 		}
// 	}

// 	common.Uint64SorterProcess(equalList)
// 	for i := len(equalList) - 1; i >= 0; i-- {
// 		tmp := equalList[i]
// 		index := candidateIndex[tmp.Name]
// 		candidates.Data[index].Rank = rank
// 		rank++
// 	}
// 	equalList = make([]common.Uint64Sort, 0)

// 	first = true
// 	for _, e := range rankSpare {
// 		if first {
// 			socre = e.Value
// 			first = false
// 			index := candidateIndex[e.Name]
// 			equalList = append(equalList, common.Uint64Sort{e.Name, uint64(index)})
// 		} else {
// 			if socre == e.Value {
// 			} else {
// 				common.Uint64SorterProcess(equalList)
// 				for i := len(equalList) - 1; i >= 0; i-- {
// 					tmp := equalList[i]
// 					index := candidateIndex[tmp.Name]
// 					candidates.Data[index].Rank = rank
// 					rank++
// 				}
// 				equalList = make([]common.Uint64Sort, 0)
// 				socre = e.Value
// 			}
// 			index := candidateIndex[e.Name]
// 			equalList = append(equalList, common.Uint64Sort{e.Name, uint64(index)})
// 		}
// 	}

// 	common.Uint64SorterProcess(equalList)
// 	for i := len(equalList) - 1; i >= 0; i-- {
// 		tmp := equalList[i]
// 		index := candidateIndex[tmp.Name]
// 		candidates.Data[index].Rank = rank
// 		rank++
// 	}
// 	equalList = make([]common.Uint64Sort, 0)

// 	first = true
// 	for _, e := range rankDie {
// 		if first {
// 			socre = e.Value
// 			first = false
// 			index := candidateIndex[e.Name]
// 			equalList = append(equalList, common.Uint64Sort{e.Name, uint64(index)})
// 		} else {
// 			if socre == e.Value {
// 			} else {
// 				common.Uint64SorterProcess(equalList)
// 				for i := len(equalList) - 1; i >= 0; i-- {
// 					tmp := equalList[i]
// 					index := candidateIndex[tmp.Name]
// 					candidates.Data[index].Rank = rank
// 					rank++
// 				}
// 				equalList = make([]common.Uint64Sort, 0)
// 				socre = e.Value
// 			}
// 			index := candidateIndex[e.Name]
// 			equalList = append(equalList, common.Uint64Sort{e.Name, uint64(index)})
// 		}
// 	}

// 	common.Uint64SorterProcess(equalList)
// 	for i := len(equalList) - 1; i >= 0; i-- {
// 		tmp := equalList[i]
// 		index := candidateIndex[tmp.Name]
// 		candidates.Data[index].Rank = rank
// 		rank++
// 	}
// }
