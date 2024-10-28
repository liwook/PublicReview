package shopservice

import "sync"

type UserLock struct {
	locks sync.Map
}

func NewUserLock() *UserLock {
	return &UserLock{}
}

var UserLockMap = NewUserLock()

func (ul *UserLock) Lock(userID int) {
	var lock *sync.Mutex
	value, ok := ul.locks.Load(userID)
	if ok {
		lock = value.(*sync.Mutex)
	} else {
		lock = &sync.Mutex{} //没有的话，就需要创建一个新锁
		ul.locks.Store(userID, lock)
	}
	lock.Lock()
}

func (ul *UserLock) Unlock(userID int) {
	value, ok := ul.locks.Load(userID)
	if ok {
		lock := value.(*sync.Mutex)
		lock.Unlock()
	}
}
