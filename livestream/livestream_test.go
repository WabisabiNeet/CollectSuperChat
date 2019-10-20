package livestream_test

import (
	"context"
	"os"
	"testing"

	"github.com/WabisabiNeet/CollectSuperChat/livestream"
	"google.golang.org/api/option"
	"google.golang.org/api/youtube/v3"
	// "github.com/WabisabiNeet/CollectSuperChat"
)

func TestGetLiveStreamID(tt *testing.T) {
	ctx := context.Background()

	apikey := os.Getenv("YOUTUBE_WATCH_LIVE_KEY")
	if apikey == "" {
		tt.Fatal("not found api key.")
	}

	ys, err := youtube.NewService(ctx, option.WithAPIKey(apikey))
	if err != nil {
		tt.Fatal(err)

	}

	// e.g. https://www.youtube.com/channel/UCp-5t9SrOQwXMU7iIjQfARg
	_, err = livestream.GetLiveStreamID(ys, "UCp-5t9SrOQwXMU7iIjQfARg")
	if err != nil {
		tt.Fatal(err)
	}
}
