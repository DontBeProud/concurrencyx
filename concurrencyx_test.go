package concurrencyx

import (
	"context"
	"github.com/go-redis/redis/v8"
	"sync"
	"testing"
	"time"
)

const (
	redisCon = "localhost:6379"
	redisPsw = ""
	redisDb  = 8
)

var (
	ctx = context.Background()
	rdb = redis.NewClient(&redis.Options{
		Addr:     redisCon,
		Password: redisPsw,
		DB:       redisDb,
	})
)

func TestConcurrencyX(t *testing.T) {
	// 模拟短信发送业务(限定每秒调用云平台接口上限次数100次)
	serviceName := "testServiceName_SendSms"
	strategy := map[Seconds]DenyThreshold{
		1: 1000,
	}
	cx, err := CreateConcurrencyX(ctx, serviceName, rdb, strategy)
	if err != nil {
		t.Error(err.Error())
		return
	}

	cnt := 0
	cntLck := sync.Mutex{}
	wg := sync.WaitGroup{}
	for i := 0; i < 2000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// 判断是否达到并发量阈值
			er := cx.Join(ctx, 10*time.Second)

			// print res
			cntLck.Lock()
			cnt += 1
			if er == nil {
				println(time.Now().UnixMilli(), cnt, "send sms success")
			} else {
				println(time.Now().UnixMilli(), cnt, "send sms fail. "+er.Error())
			}
			cntLck.Unlock()

		}()
	}
	wg.Wait()
}
