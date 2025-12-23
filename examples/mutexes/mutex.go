package mutexes

import (
	"fmt"
	"sync"

)

type bank_balance struct {
	balance int
	currency string
	mu sync.Mutex
}

var myBB = &bank_balance{balance: 1000 , currency: "tk"}


func (bb bank_balance) Add(toAdd int){
	bb.mu.Lock()
	bb.balance += toAdd
	bb.mu.Unlock()
}

func (bb bank_balance) Get() string{
	bb.mu.Lock()
	defer bb.mu.Unlock() // deferred cause the return itself return and mutex unlocking process in themselves can cause a race condition .
	// via this defer we're ensuring that the mutex is unlocked right after return statement , if we don't the function's stack frame might pop out of the stack resulting in the mutex being locked 

	return fmt.Sprintf("%d %s is your current balance" , bb.balance , bb.currency)
}