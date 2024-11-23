package main

import (
	"testing"
	"time"

	"code.cloudfoundry.org/clock/fakeclock"
	"github.com/stretchr/testify/require"
)

func Test_RequestHandler(t *testing.T) {
	type test struct {
		name      string
		durations []time.Duration
		want      []bool
		isPremium bool
	}

	tests := []test{
		{
			name:      "free user, within limit, 5s",
			isPremium: false,
			durations: []time.Duration{
				time.Second * 5,
			},
			want: []bool{true},
		},
		{
			name:      "free user, beyond limit, 20s",
			isPremium: false,
			durations: []time.Duration{
				time.Second * 20,
			},
			want: []bool{false},
		},
		{
			name:      "premium user, within free limit, 5s",
			isPremium: true,
			durations: []time.Duration{
				time.Millisecond * 5,
			},
			want: []bool{true},
		},
		{
			name:      "premium user, beyond free limit, 20s",
			isPremium: true,
			durations: []time.Duration{
				time.Millisecond * 20,
			},
			want: []bool{true},
		},
		{
			name:      "premium user, beyond free limit, 20s",
			isPremium: true,
			durations: []time.Duration{
				time.Millisecond * 20,
				time.Millisecond * 20,
				time.Millisecond * 20,
			},
			want: []bool{true, true, true},
		},
		// /// time tracking per user
		{
			name:      "free user, within limit, multiple requests, 5s",
			isPremium: false,
			durations: []time.Duration{
				time.Second * 1,
				time.Second * 2,
				time.Second * 3,
			},
			want: []bool{true, true, true},
		},
		{
			name:      "free user, beyond limit, multiple requests 14s",
			isPremium: false,
			durations: []time.Duration{
				time.Second * 5,
				time.Second * 4,
				time.Second * 5,
			},
			want: []bool{true, true, false},
		},
		{
			name:      "free user, duration same as threshold",
			isPremium: false,
			durations: []time.Duration{
				time.Second * 10,
			},
			want: []bool{true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, len(tt.want), len(tt.durations))

			fakeClock := fakeclock.NewFakeClock(time.Now())
			app := app{
				clock:     fakeClock,
				threshold: time.Second * 10,
			}
			u := User{
				IsPremium: tt.isPremium,
			}

			for i, td := range tt.durations {
				resCh := make(chan bool)

				startCh := make(chan bool)
				stopTick := make(chan bool)

				// make request
				go func() {
					res := app.HandleRequest(func() {
						startCh <- true
						app.clock.Sleep(td)
						stopTick <- true
					}, &u)
					resCh <- res
				}()

				nextStartCh := make(chan bool)

				// capture results
				go func() {
					<-startCh
					nextStartCh <- true
				}()

				// advance the clock
				go func() {
					<-nextStartCh
					ticker := time.NewTicker(time.Millisecond * 1)
				loop:
					for {
						select {
						case <-ticker.C:
							fakeClock.IncrementBySeconds(1)
						case <-stopTick:
							ticker.Stop()
							break loop
						}
					}
				}()
				res := <-resCh
				require.Equal(t, tt.want[i], res)
			}
		})
	}
}
