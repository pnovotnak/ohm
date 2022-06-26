// Package ohm is a collection of state handlers
package ohm

import (
	"github.com/pnovotnak/ohm/src/config"
	"github.com/pnovotnak/ohm/src/nextdns"
	"github.com/pnovotnak/ohm/src/types"
	"log"
	"time"
)

type Handler func(key string, allowance, cooldown, lockout time.Duration, logC chan *types.LogData) Handler

func Ready(key string, allowance, cooldown, lockout time.Duration, logC chan *types.LogData) Handler {
	log.Printf("ready: %s", key)
	_ = nextdns.SetBlock(key, false)
	for {
		select {
		case <-logC:
			return Monitoring
		}
	}
}

func Monitoring(key string, allowance, cooldown, lockout time.Duration, logC chan *types.LogData) Handler {
	log.Printf("monitoring %s", key)
	end := time.Now().Add(allowance)
	cooldownTimer := time.NewTimer(cooldown)
	defer cooldownTimer.Stop()
	for {
		select {
		case <-logC:
			if time.Now().After(end) {
				return Blocking
			} else {
				cooldownTimer.Reset(cooldown)
				log.Printf("cooldown timer reset for %s (%s left, %s lockout. Resets after %s)", key, end.Sub(time.Now()), lockout, cooldown)
			}
		case <-cooldownTimer.C:
			return Ready
		}
	}
}

func Blocking(key string, allowance, cooldown, lockout time.Duration, logC chan *types.LogData) Handler {
	log.Printf("blocking %s", key)
	_ = nextdns.SetBlock(key, true)
	lockoutTimer := time.NewTimer(lockout)
	for {
		select {
		case <-lockoutTimer.C:
			return Ready
		case <-logC:
		}
	}
}

func Run(key string, bucket *config.BlockBucket, logC chan *types.LogData) {
	fn := Ready(key, *bucket.Allowance, *bucket.Cooldown, *bucket.Lockout, logC)
	for {
		fn = fn(key, *bucket.Allowance, *bucket.Cooldown, *bucket.Lockout, logC)
	}
}
