package seqno

import (
	"github.com/garyburd/redigo/redis"
	"github.com/pkg/errors"
	"gopkg.in/redsync.v1"
	"time"
)

type RedisLocker struct {
	redSyncInstance *redsync.Redsync
}

func NewRedisLocker(poolConfig *redis.Pool) *RedisLocker{
	return &RedisLocker{
		redSyncInstance:redsync.New([]redsync.Pool{poolConfig}),
	}
}

func (r *RedisLocker) Lock(objectID string,expire time.Duration) error {
	mutex := r.redSyncInstance.NewMutex(objectID,redsync.SetExpiry(expire),redsync.SetTries(3))
	return mutex.Lock()
}

func (r *RedisLocker) LockWithTimeout(objectID string,expire time.Duration,timeout time.Duration) error {
	ch := make(chan error)
	go func(){
		mutex := r.redSyncInstance.NewMutex(objectID,redsync.SetExpiry(expire),redsync.SetTries(3))
		err :=  mutex.Lock()
		ch <- err
	}()
	select {
		case ret := <- ch:
			return ret
		case <-time.After(timeout):
			return errors.New("[SeqNo] lock object timeout:" + objectID)
	}
}



func (r *RedisLocker) Unlock(objectID string) error {
	mutex := r.redSyncInstance.NewMutex(objectID,redsync.SetTries(3))
	if mutex.Unlock() {
		return nil
	}
	return errors.New("[SeqNo] can't unlock with:" + objectID)
}

