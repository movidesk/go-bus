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
