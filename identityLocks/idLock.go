package identitylocks

import (
	"sync"

	"github.com/google/uuid"
)

type IdentityLock struct {
	id       string     // identity of process/object who has taken up the lock.
	isLocked bool       // tells whether lock has been obtained by someone
	l        sync.Mutex // lock
}

func (i *IdentityLock) Lock(id string) string {
	if id == "" {
		id = uuid.NewString()
	}
	i.l.Lock()
	i.id = id
	i.isLocked = true
	return id
}

func (i *IdentityLock) Unlock() {
	i.l.Unlock()
	i.id = ""
	i.isLocked = false
}

func (i *IdentityLock) IsLocked() (bool, string) {
	return i.isLocked, i.id
}

func NewIdentityLock() *IdentityLock {
	return &IdentityLock{
		id:       "",
		isLocked: false,
		l:        sync.Mutex{},
	}
}
