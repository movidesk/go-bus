package amqp

import "time"

func waitToBeTrue(check func() bool, d time.Duration) {
	end := time.Now().Add(d)
	for {
		if check() {
			return
		}

		if end.Before(time.Now()) {
			return
		}
	}
}

func waitForTimeout(fn func(), d time.Duration) bool {
	c := make(chan struct{})
	go func() {
		defer close(c)
		fn()
	}()
	select {
	case <-c:
		return false
	case <-time.After(d):
		return true
	}
}
