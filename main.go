package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"slices"
	"steamprofilewatcher/steam"
	"strconv"
	"strings"
	"time"
)

var (
	steamKey = flag.String("steamKey", "TOP_SECRET", "the Steam API key from https://steamcommunity.com/dev/apikey")
	steamID  = flag.Int64("steamID", 11223344556677880, "the Steam Account ID from your Steam profile page")
)

var parseMode = flag.Bool("parseMode", true, "whether to parse rather than generate log csv")

func main() {
	flag.Parse()
	if *parseMode {
		if err := parse(); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := generate(); err != nil {
			log.Fatal(err)
		}
	}
}

type Point struct {
	Date            string
	PlayTimeMinutes int
}

func parse() error {
	fp, err := os.Open(statisticFilename())
	if err != nil {
		return fmt.Errorf("open stat file: %w", err)
	}
	scanner := bufio.NewScanner(fp)
	idToName, idToPoints, err := csvScan(scanner)
	if err := outputReport(buildReport(idToName, idToPoints)); err != nil {
		return err
	}
	return nil
}

func csvScan(lineScanner *bufio.Scanner) (idToName map[string]string, idToPoints map[string][]Point, err error) {
	idToName = make(map[string]string)
	idToPoints = make(map[string][]Point)
	var doneHeader bool
	for lineScanner.Scan() {
		if !doneHeader {
			doneHeader = true
			continue
		}
		line := lineScanner.Text()
		date, stat, err := steam.ParseCSVLine(line)
		if err != nil {
			return nil, nil, fmt.Errorf("parse stat line: %w", err)
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
		return nil, nil, fmt.Errorf("read stat file: %w", lineScanner.Err())
	}
	return idToName, idToPoints, nil
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

func buildReport(idToName map[string]string, idToPoints map[string][]Point) []string {
	// Comment: now that I keep playing through days and skip period by hibernation,
	// the heartbeat on each day can be multi day sum.

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

func outputReport(lines []string) error {
	filename := filepath.Join(exeDirOrEmpty(), "report.csv")
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create report file: %w", err)
	}

	if err := writeBOMPrefix(file); err != nil {
		return err
	}
	for _, line := range lines {
		_, err := file.WriteString(line + "\n")
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

func generate() error {
	c, err := steam.NewClient(*steamKey, *steamID)
	if err != nil {
		return err
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelFunc()
	stats, err := c.GetRecentlyPlayedGameStats(ctx)
	if err != nil {
		return err
	}
	return output(stats)
}

func exeDirOrEmpty() string {
	executable, err := os.Executable()
	fmt.Println(executable)
	if err != nil {
		return ""
	}
	absolute, err := filepath.EvalSymlinks(executable)
	fmt.Println(absolute)
	if err != nil {
		return ""
	}
	return filepath.Dir(absolute)
}

func statisticFilename() string {
	return filepath.Join(exeDirOrEmpty(), "steam_profile_watch.csv")
}

func output(stats []steam.GameStat) (e error) {
	filename := statisticFilename()
	file, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND, 0)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		file, err = os.Create(filename)
		if err != nil {
			return err
		}
		if err := prepareCSVHeader(file); err != nil {
			return err
		}
	}
	defer func(file *os.File) {
		e = errors.Join(e, file.Close())
	}(file)

	now := time.Now()
	for _, stat := range stats {
		if _, err := file.WriteString(stat.CSVLine(now) + "\n"); err != nil {
			return err
		}
	}
	return nil
}

func writeBOMPrefix(w io.Writer) error {
	// Add BOM for Excel kanji print.
	_, err := w.Write([]byte("\uFEFF"))
	return err
}

func prepareCSVHeader(w io.Writer) error {
	if err := writeBOMPrefix(w); err != nil {
		return err
	}
	if _, err := w.Write([]byte(steam.GameStat{}.CSVHeader() + "\n")); err != nil {
		return err
	}
	return nil
}
