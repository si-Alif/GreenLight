package mutexes

// the goal of using RWmutex is , in normal mutex we're holding the lock in Get while what we're performing is just read operation . We want to implement something that multiple goroutines can perform read operation at the same time IF no write operation isn't being performed

import (
	"fmt"
	"sync"
)

var myBalance = &balance{amount: 5000, currency: "tk"}

type balance struct {
	amount   int
	currency string
	mu       sync.RWMutex // RWMutex instead of mutex
}

func (b *balance) Add(i int) {
	b.mu.Lock()
	b.amount += i
	b.mu.Unlock()
}

func (b *balance) Get() string {
	b.mu.RLock() // only Read lock 
	defer b.mu.RUnlock()

	return fmt.Sprintf("%d %s", b.amount, b.currency)
}

