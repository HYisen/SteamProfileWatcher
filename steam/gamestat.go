package steam

import (
	"strconv"
	"strings"
	"time"
)

type GameStat struct {
	ID                      string
	Name                    string
	PlayTimeTwoWeeksMinutes int
	PlayTimeForeverMinutes  int
}

func (gs GameStat) CSVHeader() string {
	return "EpochMilli,Date,ID,Name,PlayTimeTwoWeeksMinutes,PlayTimeForeverMinutes"
}

func (gs GameStat) CSVLine(now time.Time) string {
	fields := []string{
		strconv.FormatInt(now.UnixMilli(), 10),
		now.Format(time.DateOnly),
		gs.ID,
		gs.Name,
		strconv.Itoa(gs.PlayTimeTwoWeeksMinutes),
		strconv.Itoa(gs.PlayTimeForeverMinutes),
	}
	return strings.Join(fields, ",")
}
