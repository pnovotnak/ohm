// Package ohm is a collection of state handlers
package ohm

import (
	"github.com/pnovotnak/ohm/src/config"
	"github.com/pnovotnak/ohm/src/nextdns"
	"github.com/pnovotnak/ohm/src/types"
	"log"
	"time"
)

const MaxDuration = time.Duration(1<<63 - 62135596801)

type Handler func(key string, allowance, cooldown, lockout time.Duration, logC chan types.LogData) Handler

func durationOrMax(duration time.Duration) time.Duration {
	if duration > 0 {
		return duration
	} else {
		return MaxDuration
	}
}

func Ready(key string, _, _, _ time.Duration, logC chan types.LogData) Handler {
	resp, _ := nextdns.SetBlock(key, false)
	log.Printf("ready: %s, unblocked with status %d", key, resp.StatusCode)
	for {
		select {
		case <-logC:
			return Monitoring
		}
	}
}

func Monitoring(key string, allowance, cooldown, lockout time.Duration, logC chan types.LogData) Handler {
	log.Printf("monitoring: %s (cooldown: %s, lockout: %s)", key, cooldown, lockout)
	end := time.Now().Add(allowance)

	// if a cooldown timer is provided, use that
	cooldownTimer := time.NewTimer(durationOrMax(cooldown))
	defer cooldownTimer.Stop()

	var sessionTimer *time.Timer
	if cooldown.Milliseconds() > 0 {
		// we need the session timer to never fire if we're using cooldown
		sessionTimer = time.NewTimer(MaxDuration)
	} else {
		sessionTimer = time.NewTimer(allowance)
	}
	defer sessionTimer.Stop()

	for {
		select {
		case <-logC:
			if time.Now().After(end) {
				return Blocking
			} else if cooldown.Milliseconds() > 0 {
				cooldownTimer.Reset(durationOrMax(cooldown))
				log.Printf("cooldown timer reset for %s (%s left in session, %s lockout. Resets after %s)", key, end.Sub(time.Now()), lockout, cooldown)
			}
		case <-cooldownTimer.C:
			return Ready
		case <-sessionTimer.C:
			return Blocking
		}
	}
}

func Blocking(key string, _, _, lockout time.Duration, logC chan types.LogData) Handler {
	resp, _ := nextdns.SetBlock(key, true)
	lockoutTimer := time.NewTimer(lockout)
	log.Printf("blocking: %s, blocked with status %d", key, resp.StatusCode)
	for {
		select {
		case <-lockoutTimer.C:
			return Ready
		case <-logC:
		}
	}
}

func Run(key string, bucket *config.BlockBucket, logC chan types.LogData) {
	fn := Ready(key, bucket.Allowance, bucket.Cooldown, bucket.Lockout, logC)
	for {
		fn = fn(key, bucket.Allowance, bucket.Cooldown, bucket.Lockout, logC)
	}
}
