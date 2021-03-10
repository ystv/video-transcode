package task

import (
	"strconv"
	"strings"
)

// Stats represents statistics on the current encode job
type Stats struct {
	Duration   int    `json:"duration"`
	Percentage int    `json:"percentage"`
	Frame      int    `json:"frame"`
	FPS        int    `json:"fps"`
	Bitrate    string `json:"bitrate"`
	Size       string `json:"size"`
	Time       string `json:"time"`
}

func getStats(s *Stats, res string) bool {
	durIdx := strings.Index(res, "Duration")
	// Checking if we've got a "Duration",
	// we need this so we can determine the ETA
	if durIdx >= 0 {
		dur := res[durIdx+10:]
		// handle if video doesn't have a length i.e. live
		if dur[0:3] == "N/A" {
			s.Duration = -1
			s.Percentage = 100
			return true
		} else if len(dur) > 8 {
			dur = dur[0:8]

			s.Duration = durToSec(dur)
			return true
		}
	}
	// FFmpeg should give us a duration on startup,
	// so we kill here in the event that didn't happen.
	if s.Duration == 0 {
		return false
	}

	frameIdx := strings.LastIndex(res, "frame=")
	fpsIdx := strings.LastIndex(res, "fps=")
	bitrateIdx := strings.LastIndex(res, "bitrate=")
	sizeIdx := strings.LastIndex(res, "size=")
	timeIdx := strings.Index(res, "time=")

	if timeIdx >= 0 {
		// From this point on it should be outputting normal encode stdout,
		// which we'll want to parse.

		frame := strings.Fields(res[frameIdx+6:])
		fps := strings.Fields(res[fpsIdx+4:])
		bitrate := strings.Fields(res[bitrateIdx+8:])
		size := strings.Fields(res[sizeIdx+5:])
		time := res[timeIdx+5:]

		if len(time) > 8 {
			time = time[0:8]
			sec := durToSec(time)
			per := (sec * 100) / s.Duration
			if s.Percentage != per {
				// set percentage to 100% when we don't have a length
				if s.Duration != -1 {
					s.Percentage = per
				}
				// Just doing to reuse this int variable for each item
				integer, _ := strconv.Atoi(frame[0])
				s.Frame = integer
				integer, _ = strconv.Atoi(fps[0])
				s.FPS = integer
				s.Bitrate = bitrate[0]
				s.Size = size[0]
				s.Time = time
			}
			return true
		}
	}
	return false
}

func durToSec(dur string) (sec int) {
	// So we're kind of limiting our videos to 24h which isn't ideal
	// shouldn't crash the application hopefully XD
	durAry := strings.Split(dur, ":")
	if len(durAry) != 3 {
		return
	}
	hr, _ := strconv.Atoi(durAry[0])
	sec = hr * (60 * 60)
	min, _ := strconv.Atoi(durAry[1])
	sec += min * (60)
	second, _ := strconv.Atoi(durAry[2])
	sec += second
	return
}
