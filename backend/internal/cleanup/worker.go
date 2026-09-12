package cleanup

import (
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	"github.com/Shubham-rawat0/temp-mail/SwiftIndbox/backend/internal/repository/postgres"
)


type Scheduler struct{
	timer 			*time.Timer
	running  		bool
	defaultInterval time.Duration
	startDelay 		time.Duration
	q 				*postgres.Queries	
}

const (
	defaultIntervalFallback = 6 * time.Hour
	maxStartDelay            = 5 * time.Minute
	jitterPct                = 0.10
	minDelay                 = time.Minute
)

func NewScheduler(q *postgres.Queries) *Scheduler{
	defaultInterval := getDurationEnv(
		"CLEANUP_INTERVAL_MS",
		defaultIntervalFallback,
	)

	startDelay := getDurationEnv(
		"CLEANUP_START_DELAY_MS",
		min(defaultInterval/2, maxStartDelay),
	)

	return &Scheduler{
		defaultInterval: defaultInterval,
		startDelay:      startDelay,
		q:				 q,
	}
}

func (s *Scheduler) Start(){
	if s.running {
		log.Println("[SCHEDULER] Already running")
		return
	}
	s.running = true
	log.Printf(
		"[SCHEDULER] Enabled. interval≈%v, start in %v",
		s.defaultInterval,
		s.startDelay,
	)
	s.scheduleNext(s.startDelay)
}

func (s *Scheduler) Stop(){
	if s.timer!=nil{
		s.timer.Stop()
		s.timer=nil
	}
	s.running = false
	log.Println("[SCHEDULER] Stopped")
}

func (s *Scheduler) GetStatus() map[string]bool {
	return map[string]bool{
		"isRunning": s.running,
	}
}

func (s *Scheduler) scheduleNext(baseMs time.Duration){
	if !s.running{
		return
	}
	jitterFactor := 1 + (rand.Float64()*2*jitterPct - jitterPct)
	delay := time.Duration(float64(baseMs) * jitterFactor)
	delay=max(delay,minDelay)
	
	time.AfterFunc(delay,s.tick)
}

func (s *Scheduler) tick(){
	if !s.running{
		return
	}
	res,err:=CleanUpExpired(s.q)
	if err!=nil{
		s.scheduleNext(s.defaultInterval)
		return
	}

	var next time.Duration
	if res > 0 {
		next = min(s.defaultInterval, 30*time.Minute)
	} else {
		next = s.defaultInterval
	}
	s.scheduleNext(next)
}

func getDurationEnv(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)

	if value == "" {
		return fallback
	}

	ms, err := strconv.Atoi(value)
	if err != nil {
		log.Printf(
			"[SCHEDULER] Invalid %s=%q, using default %v",
			key,
			value,
			fallback,
		)
		return fallback
	}

	return time.Duration(ms) * time.Millisecond
}