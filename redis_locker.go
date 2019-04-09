package seqno

import (
	"github.com/garyburd/redigo/redis"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"gopkg.in/redsync.v1"
	"time"
)

var mutexCache *cache.Cache

func init() {
	mutexCache = cache.New(60 * time.Second,120 * time.Second)
}

type RedisLocker struct {
	redSyncInstance *redsync.Redsync
}

func delayFunc(i int) time.Duration {
	return time.Duration(i*50) * time.Millisecond
}

func NewRedisLocker(poolConfig *redis.Pool) *RedisLocker {
	return &RedisLocker{
		redSyncInstance: redsync.New([]redsync.Pool{poolConfig}),
	}
}

func (r *RedisLocker) Lock(objectID string, expire time.Duration) error {
	mutex := r.redSyncInstance.NewMutex(objectID, redsync.SetExpiry(expire), redsync.SetTries(3))
	mutexCache.Set(objectID,mutex,expire)
	return mutex.Lock()
}

func (r *RedisLocker) LockWithTimeout(objectID string, expire time.Duration, timeout time.Duration) error {
	ch := make(chan error)
	go func() {
		mutex := r.redSyncInstance.NewMutex(objectID, redsync.SetExpiry(expire), redsync.SetTries(3), redsync.SetRetryDelay(time.Duration(timeout/3)))
		mutexCache.Set(objectID,mutex,expire)
		err := mutex.Lock()
		ch <- err
	}()
	select {
	case ret := <-ch:
		return ret
	case <-time.After(timeout):
		return errors.New("[SeqNo] lock object timeout:" + objectID)
	}
}

func (r *RedisLocker) Unlock(objectID string) error {
	var mutex *redsync.Mutex
	oldMutex ,ok:= mutexCache.Get(objectID)
	if ok {
		mutex = oldMutex.(*redsync.Mutex)
	}else{
		mutex = r.redSyncInstance.NewMutex(objectID, redsync.SetTries(3))
	}
	if mutex.Unlock() {
		mutexCache.Delete(objectID)
		return nil
	}
	return errors.New("[SeqNo] can't unlock with:" + objectID)
}