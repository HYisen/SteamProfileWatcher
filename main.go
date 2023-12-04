package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"steamprofilewatcher/steam"
)

var (
	steamKey = flag.String("steamKey", "TOP_SECRET", "the Steam API key from https://steamcommunity.com/dev/apikey")
	steamID  = flag.Int64("steamID", 11223344556677880, "the Steam Account ID from your Steam profile page ")
)

func main() {
	flag.Parse()
	if err := handle(); err != nil {
		log.Fatal(err)
	}
}

func handle() error {
	c, err := steam.NewClient(*steamKey, *steamID)
	if err != nil {
		return err
	}
	stats, err := c.GetRecentlyPlayedGameStats(context.TODO())
	if err != nil {
		return err
	}
	for _, stat := range stats {
		fmt.Printf("%+v\n", stat)
	}
	return nil
}
