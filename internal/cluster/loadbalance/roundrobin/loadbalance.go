package roundrobin

import (
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/luojinbo008/gost/internal/cluster/loadbalance"
	"github.com/luojinbo008/gost/internal/protocol"
)

const (
	Complete = 0
	Updating = 1
)

var (
	methodWeightMap sync.Map          // [string]invokers
	state           = int32(Complete) // update lock acquired ?
	recyclePeriod   = 60 * time.Second.Nanoseconds()
)

func init() {
	// extension.SetLoadbalance(constant.LoadBalanceKeyRoundRobin, NewRRLoadBalance)
}

type rrLoadBalance struct{}

// NewRRLoadBalance returns a round robin load balance
//
// Use the weight's common advisory to determine round robin ratio
func NewRRLoadBalance() loadbalance.LoadBalance {
	return &rrLoadBalance{}
}

// Select gets invoker based on round robin load balancing strategy
func (lb *rrLoadBalance) Select(invokers []protocol.Invoker, invocation protocol.Invocation) protocol.Invoker {
	count := len(invokers)
	if count == 0 {
		return nil
	}
	if count == 1 {
		return invokers[0]
	}

	key := invokers[0].GetURL().Path + "." + invocation.MethodName()
	cache, _ := methodWeightMap.LoadOrStore(key, &cachedInvokers{})
	cachedInvokers := cache.(*cachedInvokers)

	var (
		clean               = false
		totalWeight         = int64(0)
		maxCurrentWeight    = int64(math.MinInt64)
		now                 = time.Now()
		selectedInvoker     protocol.Invoker
		selectedWeightRobin *weightedRoundRobin
	)

	for _, invoker := range invokers {
		// weight := loadbalance.GetWeight(invoker, invocation)
		// 默认权重都是1
		weight := int64(1)
		if weight < 0 {
			weight = 0
		}

		identifier := invoker.GetURL().Key()
		loaded, found := cachedInvokers.LoadOrStore(identifier, &weightedRoundRobin{weight: weight})
		weightRobin := loaded.(*weightedRoundRobin)
		if !found {
			clean = true
		}

		if weightRobin.Weight() != weight {
			weightRobin.setWeight(weight)
		}

		currentWeight := weightRobin.increaseCurrent()
		weightRobin.lastUpdate = &now

		if currentWeight > maxCurrentWeight {
			maxCurrentWeight = currentWeight
			selectedInvoker = invoker
			selectedWeightRobin = weightRobin
		}
		totalWeight += weight
	}

	cleanIfRequired(clean, cachedInvokers, &now)

	if selectedWeightRobin != nil {
		selectedWeightRobin.Current(totalWeight)
		return selectedInvoker
	}

	// should never happen
	return invokers[0]
}

func cleanIfRequired(clean bool, invokers *cachedInvokers, now *time.Time) {
	if clean && atomic.CompareAndSwapInt32(&state, Complete, Updating) {
		defer atomic.CompareAndSwapInt32(&state, Updating, Complete)
		invokers.Range(func(identify, robin interface{}) bool {
			weightedRoundRobin := robin.(*weightedRoundRobin)
			elapsed := now.Sub(*weightedRoundRobin.lastUpdate).Nanoseconds()
			if elapsed > recyclePeriod {
				invokers.Delete(identify)
			}
			return true
		})
	}
}

// Record the weight of the invoker
type weightedRoundRobin struct {
	weight     int64
	current    int64
	lastUpdate *time.Time
}

func (robin *weightedRoundRobin) Weight() int64 {
	return atomic.LoadInt64(&robin.weight)
}

func (robin *weightedRoundRobin) setWeight(weight int64) {
	robin.weight = weight
	robin.current = 0
}

func (robin *weightedRoundRobin) increaseCurrent() int64 {
	return atomic.AddInt64(&robin.current, robin.weight)
}

func (robin *weightedRoundRobin) Current(delta int64) {
	atomic.AddInt64(&robin.current, -1*delta)
}

type cachedInvokers struct {
	sync.Map /*[string]weightedRoundRobin*/
}
