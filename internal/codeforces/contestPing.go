package codeforces

import "time"

const pingTime int = 1 * 3600 // 1 hour

// Start goroutine that checks whether it should issue a ping for upcoming contests
func (man *manager) startContestPingCheck() {
	go func() {
		for {
			time.Sleep(1 * time.Minute)

			man.checkContestPing()
		}
	}()
}

func (man *manager) checkContestPing() {
	for _, contest := range man.upcomingContests {
		if contest.StartTimeSeconds - int(time.Now().Unix()) <= pingTime && !contest.Pinged {
			contestPing(&contest)
		}
	}
}

func contestPing(contest *contest) {
	contest.Pinged = true
}
