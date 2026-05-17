package rest

import (
	"sync"
	"time"
)

const (
	defaultRateLimitRate  = 5.0
	defaultRateLimitBurst = 10.0
)

// RateLimitResult describes the outcome of a rate-limit check.
type RateLimitResult int

const (
	// RateLimitAllowed means the submission is within limits.
	RateLimitAllowed RateLimitResult = iota
	// RateLimitFirstBreach means the team just exceeded the limit for the first
	// time; callers should emit a "blocked" alert event and reject the request.
	RateLimitFirstBreach
	// RateLimitBlocked means the team is still within a grace-period block;
	// callers should reject the request without re-emitting an event.
	RateLimitBlocked
	// RateLimitUnblocked means the grace period just expired and this request
	// is allowed; callers should emit an "unblocked" alert event and proceed.
	RateLimitUnblocked
)

type teamRateLimitState struct {
	mu           sync.Mutex
	tokens       float64
	lastRefil    time.Time
	blockedUntil time.Time
}

// check applies token-bucket logic and returns the appropriate RateLimitResult.
// rate, burst, and gracePeriodMinutes are read from config on every call so
// live config reloads take effect immediately.
func (s *teamRateLimitState) check(now time.Time, rate, burst, gracePeriodMinutes float64) RateLimitResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If still within the grace-period block, keep rejecting.
	if now.Before(s.blockedUntil) {
		return RateLimitBlocked
	}

	// Grace period just expired — clear the block and signal the unblock.
	wasBlocked := !s.blockedUntil.IsZero()
	s.blockedUntil = time.Time{}

	// Refill tokens based on elapsed time.
	elapsed := now.Sub(s.lastRefil).Seconds()
	s.tokens += elapsed * rate
	if s.tokens > burst {
		s.tokens = burst
	}

	s.lastRefil = now

	if s.tokens >= 1 {
		s.tokens--

		if wasBlocked {
			return RateLimitUnblocked
		}

		return RateLimitAllowed
	}

	// Bucket still empty after grace period — start a new block.
	s.blockedUntil = now.Add(time.Duration(gracePeriodMinutes * float64(time.Minute)))

	return RateLimitFirstBreach
}

type teamRateLimiter struct {
	mu     sync.Mutex
	states map[int64]*teamRateLimitState
}

func newTeamRateLimiter() *teamRateLimiter {
	return &teamRateLimiter{
		states: make(map[int64]*teamRateLimitState),
	}
}

// check records a flag submission for teamID and returns the rate-limit outcome.
// Returns RateLimitAllowed when rate limiting is disabled (gracePeriod is zero
// or negative). Rate defaults to defaultRateLimitRate and burst defaults to
// defaultRateLimitBurst when not set in config.
func (rl *teamRateLimiter) check(teamID int64, rate, burst, gracePeriodMinutes float64) RateLimitResult {
	if gracePeriodMinutes <= 0 {
		return RateLimitAllowed
	}

	if rate <= 0 {
		rate = defaultRateLimitRate
	}

	if burst <= 0 {
		burst = defaultRateLimitBurst
	}

	rl.mu.Lock()
	state, ok := rl.states[teamID]
	if !ok {
		state = &teamRateLimitState{
			tokens:    burst,
			lastRefil: time.Now(),
		}
		rl.states[teamID] = state
	}
	rl.mu.Unlock()

	return state.check(time.Now(), rate, burst, gracePeriodMinutes)
}
