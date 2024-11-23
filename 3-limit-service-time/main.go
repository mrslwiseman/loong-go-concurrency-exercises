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
	"sync/atomic"
	"time"

	"code.cloudfoundry.org/clock"
)

// User defines the UserModel. Use this to check whether a User is a
// Premium user or not
type User struct {
	ID        int
	IsPremium bool
	TimeUsed  atomic.Uint64
}

func (u *User) TrackUsage(duration uint64) uint64 {
	return u.TimeUsed.Add(1)
}

type app struct {
	clock     clock.Clock
	threshold time.Duration
}

func New(threshold time.Duration) app {
	clock := clock.NewClock()
	return app{
		clock:     clock,
		threshold: threshold,
	}
}

// HandleRequest runs the processes requested by users. Returns false
// if process had to be killed
func (a *app) HandleRequest(process func(), u *User) bool {
	resCh := make(chan bool)
	startCh := make(chan bool)

	// do the work
	go func() {
		startCh <- true
		if u.IsPremium {
			// dont care about duration for premium users
			resCh <- true
			process()
		} else {
			process()
			resCh <- true
		}
	}()

	// in case same as threshold, give them benefit
	// TODO: Track by user
	tick := a.clock.NewTicker(time.Second)
	defer tick.Stop()
	<-startCh
	for {
		select {
		case <-tick.C():
			used := u.TimeUsed.Add(1)
			if used <= uint64(a.threshold.Seconds()) {
				continue
			}
			return false
		case <-resCh:
			return true
		}

	}
}

func main() {
	RunMockServer()

}
