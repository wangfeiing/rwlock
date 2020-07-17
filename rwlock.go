package rwlock

type RWLock struct {
	shaHashID *string
	lockKey   string
	uniqID    string
}

func New(key string) *RWLock {
	return &RWLock{
		shaHashID: shaHashID,
		lockKey:   key,
		uniqID:    "random_str_",
	}
}

func (l *RWLock) Lock() {

}

func (l *RWLock) Unlock() {

}

func (l *RWLock) RLock() {

}

func (l *RWLock) RUnlock() {

}
