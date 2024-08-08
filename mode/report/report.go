package report

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"slices"
	"steamprofilewatcher/helpers"
	"steamprofilewatcher/steam"
	"strconv"
	"strings"
)

type Point struct {
	Date            string
	PlayTimeMinutes int
}

// Data are messages generated by CSVScan and then shall be used by Build.
// Users are not supposed to creates it by themselves, just pass it.
// Its members are not private, so that users can patch or even create(although not encouraged) before Build use it.
// Otherwise, I could use OO design and seal it in receiver rather than pass between FP functions.
type Data struct {
	IDToName   map[string]string
	IDToPoints map[string][]Point
}

func CSVScan(lineScanner *bufio.Scanner) (Data, error) {
	idToName := make(map[string]string)
	idToPoints := make(map[string][]Point)
	var doneHeader bool
	for lineScanner.Scan() {
		if !doneHeader {
			doneHeader = true
			continue
		}
		line := lineScanner.Text()
		date, stat, err := steam.ParseCSVLine(line)
		if err != nil {
			return Data{}, fmt.Errorf("parse stat line: %w", err)
		}

		if name, ok := idToName[stat.ID]; ok {
			if name != stat.Name {
				log.Printf("shifted name on %v from %v to %v", stat.ID, name, stat.Name)
			}
		}
		idToName[stat.ID] = stat.Name
		idToPoints[stat.ID] = append(idToPoints[stat.ID], Point{
			Date:            date,
			PlayTimeMinutes: stat.PlayTimeForeverMinutes,
		})
	}
	if lineScanner.Err() != nil {
		return Data{}, fmt.Errorf("read stat file: %w", lineScanner.Err())
	}
	return Data{
		IDToName:   idToName,
		IDToPoints: idToPoints,
	}, nil
}

// mallocByDate fills each date in idToPoints in returned map.
// only Date info are used in input idToPoints.
func mallocByDate(idToPoints map[string][]Point) (dateToIDToPlayTimeMinutesDelta map[string]map[string]int) {
	seen := make(map[string]bool)
	for _, points := range idToPoints {
		for _, point := range points {
			seen[point.Date] = true
		}
	}

	dateToIDToPlayTimeMinutesDelta = make(map[string]map[string]int)
	for date := range seen {
		dateToIDToPlayTimeMinutesDelta[date] = make(map[string]int)
	}
	return dateToIDToPlayTimeMinutesDelta
}

func Build(data Data) []string {
	// Comment: now that I keep playing through days and skip period by hibernation,
	// the heartbeat on each day can be multi day sum.

	idToName := data.IDToName
	idToPoints := data.IDToPoints
	dateToIDToPlayTimeMinutesDelta := mallocByDate(idToPoints)
	for id, points := range idToPoints {
		slices.SortFunc(points, func(lhs, rhs Point) int {
			return strings.Compare(lhs.Date, rhs.Date) // As time.DateOnly format fit it.
		})
		dateToIDToPlayTimeMinutesDelta[points[0].Date][id] = 0
		for i := 1; i < len(points); i++ {
			delta := points[i].PlayTimeMinutes - points[i-1].PlayTimeMinutes
			dateToIDToPlayTimeMinutesDelta[points[i].Date][id] += delta
		}
	}

	var ret []string
	ret = append(ret, csvHeader(idToName))
	var keys []string
	for date, _ := range dateToIDToPlayTimeMinutesDelta {
		keys = append(keys, date)
	}
	slices.SortFunc(keys, strings.Compare)
	for _, date := range keys {
		var fields []string
		fields = append(fields, date)
		for id := range idToName {
			fields = append(fields, strconv.Itoa(dateToIDToPlayTimeMinutesDelta[date][id]))
		}
		ret = append(ret, strings.Join(fields, ","))
	}
	return ret
}

func Output(filename string, lines []string) error {
	fp, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create report file: %w", err)
	}

	if err := helpers.WriteBOMPrefix(fp); err != nil {
		return err
	}
	for _, line := range lines {
		_, err := fp.WriteString(line + "\n")
		if err != nil {
			return fmt.Errorf("write line to report file: %w", err)
		}
	}
	return nil
}

func csvHeader(idToName map[string]string) string {
	var fields []string
	fields = append(fields, "date")
	for id, name := range idToName {
		fields = append(fields, fmt.Sprintf("\"[%v]%v\"", id, name)) // enclose to allow comma in name
	}
	return strings.Join(fields, ",")
}