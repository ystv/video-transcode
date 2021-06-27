package state

import "time"

const SHORT_EXPIRY = time.Duration(2*24) * time.Hour
const LONG_EXPIRY = time.Duration(7*24) * time.Hour

func (h *StateHandler) Tidier() {
	// The Previously Mentioned Henry-Hoover Function

	for {
		// 1. Do a Tidying Pass
		for key, val := range h.Jobs {
			if fsi, ok := val.(FullStatusIndicator); ok {
				if fsi.Time.Add(SHORT_EXPIRY).After(time.Now()) {
					h.Jobs[key] = ShortStatusIndicator{
						JobID:           fsi.JobID,
						FailureMode:     fsi.FailureMode,
						Summary:         fsi.Summary,
						Time:            fsi.Time,
						FullExpiredTime: time.Now(),
					}
				}
				continue
			}

			if ssi, ok := val.(ShortStatusIndicator); ok {
				if ssi.Time.Add(LONG_EXPIRY).After(time.Now()) {
					delete(h.Jobs, key)
				}
			}
		}

		// 2. Delay
		time.Sleep(time.Duration(5) * time.Minute)

		// TODO: Stopping Call
	}
}
