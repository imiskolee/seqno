package seqno

import "time"

type Locker interface{
	Lock(objectID string,expire time.Duration) error
	LockWithTimeout(objectID string,expire time.Duration,waitTimeout time.Duration) error
	Unlock(objectID string) error
}




