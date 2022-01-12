package concurrencyx

import "errors"

var (
	ErrorReachConcurrencyThreshold = errors.New("concurrencyx error: Concurrency threshold reached")
	ErrorInvalidServiceUniqueName  = errors.New("concurrencyx error: Service unique name is invalid")
	ErrorInvalidRedisObject        = errors.New("concurrencyx error: The redis is nil")
	ErrorRedisCanNotConnect        = errors.New("concurrencyx error: Can not connect redis")
	ErrorCreateLockCreator         = errors.New("concurrencyx error: Create lock creator fail")
	ErrorSecondsFiledTooLarge      = errors.New("concurrencyx error: The value of the seconds field of the strategy is too large")
	ErrorThresholdTooSmall         = errors.New("concurrencyx error: The value of the threshold field of the strategy is too small")
)
