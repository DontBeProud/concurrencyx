package concurrencyx

import (
	"context"
	"github.com/DontBeProud/godislock"
	"github.com/go-redis/redis/v8"
)

const (
	lockNamePrefix             = "ConcurrencyX_Lock_"
	secondLevelQueueNamePrefix = "ConcurrencyX_SLQ_"
	recordPrefix               = "ConcurrencyX_R_"
)

func createConcurrencyX(ctx context.Context, srvUniqueName string, rDb *redis.Client, denyStrategy map[Seconds]DenyThreshold) (*ConcurrencyX, error) {
	if len(srvUniqueName) == 0 {
		return nil, ErrorInvalidServiceUniqueName
	}

	if rDb == nil {
		return nil, ErrorInvalidRedisObject
	}

	if _, err := rDb.Ping(context.TODO()).Result(); err != nil {
		return nil, ErrorRedisCanNotConnect
	}

	lockCreator, err := godislock.CreateRedisLockCreator(ctx, lockName(srvUniqueName), rDb)
	if err != nil {
		return nil, ErrorCreateLockCreator
	}

	strategyTable, err := createDenyStrategyTable(denyStrategy)
	if err != nil {
		return nil, err
	}

	return &ConcurrencyX{
		rDb:                             rDb,
		srvUniqueName:                   srvUniqueName,
		lockCreator:                     lockCreator,
		DenyStrategyManagementInterface: strategyTable,
		slqName:                         secondLevelQueueName(srvUniqueName),
	}, nil
}

func lockName(srvUniqueName string) string {
	return lockNamePrefix + srvUniqueName
}

func secondLevelQueueName(srvUniqueName string) string {
	return secondLevelQueueNamePrefix + srvUniqueName
}
