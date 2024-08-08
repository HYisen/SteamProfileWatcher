package steam

import (
	"fmt"
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

func ParseCSVLine(line string) (timestamp time.Time, stat GameStat, err error) {
	// If there are comma , in name, which is possible.
	// then csv.NewReader(fp).ReadAll() would say
	// record on line 19: wrong number of fields
	// So I have to implement the csv parser manually here.
	record := strings.Split(line, ",")
	epochMilli, err := strconv.ParseInt(record[0], 10, 64)
	if err != nil {
		return time.Time{}, GameStat{}, fmt.Errorf("parse timestamp: %w", err)
	}
	timestamp = time.UnixMilli(epochMilli)
	stat.ID = record[2]
	stat.Name = strings.Join(record[3:len(record)-2], ",")
	stat.PlayTimeTwoWeeksMinutes, err = strconv.Atoi(record[len(record)-2])
	if err != nil {
		return time.Time{}, GameStat{}, fmt.Errorf("parse PlayTimeTwoWeeksMinutes: %w", err)
	}
	stat.PlayTimeForeverMinutes, err = strconv.Atoi(record[len(record)-1])
	if err != nil {
		return time.Time{}, GameStat{}, fmt.Errorf("parse PlayTimeForeverMinutes: %w", err)
	}
	return timestamp, stat, nil
}
