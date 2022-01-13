package concurrencyx

import (
	"context"
	"github.com/DontBeProud/godislock"
	"github.com/go-basic/uuid"
	"github.com/go-redis/redis/v8"
	"strconv"
	"sync"
	"time"
)

const (
	// TODO: 本模块主要用于高并发场景下的并发控制，因此v0.x支持的最大时间周期范围为300秒
	secondsLimit = 300
)

type InterfaceConcurrencyX interface {
	Join(ctx context.Context, waitTimeOut time.Duration) error
}

type Seconds uint // 0 < Seconds <= secondsLimit
type DenyThreshold uint

type DenyStrategyManagementInterface interface {
	AddStrategy(s Seconds, threshold DenyThreshold) error
	DelStrategy(s Seconds)
	OutPut() *map[Seconds]DenyThreshold
}

type ConcurrencyX struct {
	lockCreator   *godislock.RedisLockCreator
	rDb           *redis.Client
	srvUniqueName string
	slqName       string
	DenyStrategyManagementInterface
	InterfaceConcurrencyX
}

func CreateConcurrencyX(ctx context.Context, srvUniqueName string, rDb *redis.Client, denyStrategy map[Seconds]DenyThreshold) (*ConcurrencyX, error) {
	return createConcurrencyX(ctx, srvUniqueName, rDb, denyStrategy)
}

func (x ConcurrencyX) Join(ctx context.Context, waitTimeOut time.Duration) error {
	resChan := make(chan error)
	continueChan := make(chan bool)
	timeOutClock := time.Tick(waitTimeOut)

	go func() {
		for{
			resChan <- x._join(ctx, waitTimeOut)
			if _, ok := <- continueChan; !ok{
				break
			}
		}
	}()

	for{
		select {
		case <- timeOutClock:
			close(continueChan)
			return ErrorWaitTimeOut
		case err := <- resChan:
			// retry
			if err == nil || err.Error() != ErrorReachConcurrencyThreshold.Error(){
				close(continueChan)
				return err
			}
			continueChan <- true
		}
	}
}

func (x ConcurrencyX) _join(ctx context.Context, waitTimeOut time.Duration) error {
	// get distributed lock
	lck, err := x.lockCreator.Acquire(ctx, 10*time.Second, waitTimeOut)
	if err != nil {
		return err
	}
	go lck.AutoRefresh(ctx)
	defer lck.Release(ctx)

	// pass or deny
	strategy := x.OutPut()
	for second, threshold := range *strategy {
		d := time.Duration(-1*int(second)) * time.Second
		cur := time.Now()

		cnt, err := x.rDb.ZCount(ctx, x.slqName, strconv.Itoa(int(cur.Add(d).UnixNano())), strconv.Itoa(int(cur.UnixNano()))).Result()
		if err != nil {
			return err
		}
		if cnt >= int64(threshold) {
			return ErrorReachConcurrencyThreshold
		}
	}

	// TODO: 改用lua脚本, 减少redis接口调用次数

	//// 23分之1的概率触发清理过期数据的任务
	//cur := time.Now()
	if time.Now().UnixMilli()%23 == 0 {
		_d := time.Duration(-600) * time.Second
		wg := sync.WaitGroup{}
		wg.Add(1)
		defer wg.Wait()
		go func() {
			defer wg.Done()
			x.rDb.ZRemRangeByScore(ctx, x.slqName, "0", strconv.Itoa(int(time.Now().Add(_d).UnixNano())))
		}()
	}

	// record
	n := time.Now().UnixNano()
	return x.rDb.ZAddNX(ctx, x.slqName, &redis.Z{
		Score:  float64(n),
		Member: recordPrefix + strconv.Itoa(int(n)) + uuid.New(),
	}).Err()
}
