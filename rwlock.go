package rwlock

type RWLock struct {
	shaHashID *string
	lockKey   string
}

func New(key string) *RWLock {
	return &RWLock{
		shaHashID: shaHashID,
		lockKey:   key,
	}
}

func (l *RWLock) Lock() {

}
