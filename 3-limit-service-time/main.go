//////////////////////////////////////////////////////////////////////
//
// Your video processing service has a freemium model. Everyone has 10
// sec of free processing time on your service. After that, the
// service will kill your process, unless you are a paid premium user.
//
// Beginner Level: 10s max per request
// Advanced Level: 10s max per user (accumulated)
//

package main

import (
	"context"
	"sync"
	"time"
)

// User defines the UserModel. Use this to check whether a User is a
// Premium user or not
type User struct {
	sync.Mutex
	ID        int
	IsPremium bool
	TimeUsed  int64 // in seconds
}

// HandleRequest runs the processes requested by users. Returns false
// if process had to be killed
func HandleRequest(process func(), u *User) bool {
	ctx, cancel := context.WithCancel(context.Background())
	var done = make(chan bool)

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				start := time.Now()
				process()
				timeElapsed := time.Since(start)
				u.Lock()
				u.TimeUsed += int64(timeElapsed.Seconds())
				u.Unlock()
				done <- true
				return
			}
		}
	}(ctx)

	for {
		select {
		case <-done:
			cancel()
			return true
		default:
			u.Lock()
			if !u.IsPremium && u.TimeUsed >= 10 {
				cancel()
				return false
			}
			u.Unlock()
			time.Sleep(10 * time.Millisecond)
		}
	}
}

func main() {
	RunMockServer()
}
