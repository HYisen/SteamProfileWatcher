package steam

import (
	"context"
	"fmt"
	"github.com/leighmacdonald/steamid/v3/steamid"
	"github.com/leighmacdonald/steamweb/v2"
	"os"
	"strconv"
	"sync/atomic"
)

type Client struct {
	sid steamid.SID64
}

var created atomic.Bool

func NewClient(apiToken string, accountID int64) (*Client, error) {
	// This procedure is simulating the upstream behavior, better keep synchronized.
	const envKey = "STEAM_TOKEN"
	if v, ok := os.LookupEnv(envKey); ok && v != "" {
		return nil, fmt.Errorf("exist env %s=%s", envKey, v)
	}

	// Limited by upstream, I would have to prevent multiple instance here first.
	if ok := created.CompareAndSwap(false, true); !ok {
		return nil, fmt.Errorf("multiple instance not supported")
	}

	if err := steamweb.SetKey(apiToken); err != nil {
		return nil, err
	}
	return &Client{sid: steamid.New(accountID)}, nil
}

func (c *Client) GetRecentlyPlayedGameStats(ctx context.Context) ([]GameStat, error) {
	games, err := steamweb.GetRecentlyPlayedGames(ctx, c.sid)
	if err != nil {
		return nil, err
	}

	var ret []GameStat
	for _, game := range games {
		ret = append(ret, GameStat{
			ID:                      strconv.FormatUint(uint64(game.AppID), 10),
			Name:                    game.Name,
			PlayTimeTwoWeeksMinutes: game.Playtime2Weeks,
			PlayTimeForeverMinutes:  game.PlaytimeForever,
		})
	}
	return ret, nil
}
