package subscriber

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/MangataL/BangumiBuddy/internal/meta"
)

func TestSubscriber_HandleEpisodeTransferred(t *testing.T) {
	const subscriptionID = "sub-1"
	ctx := context.Background()
	refreshErr := errors.New("tmdb unavailable")

	testCases := []struct {
		name    string
		episode int
		fake    func(t *testing.T) *Subscriber
	}{
		{
			name:    "when episode below total then refreshes tmdb total",
			episode: 5,
			fake: func(t *testing.T) *Subscriber {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				repo := NewMockRepository(ctrl)
				parser := meta.NewMockParser(ctrl)
				bangumi := Bangumi{SubscriptionID: subscriptionID, TMDBID: 95231, Season: 2, EpisodeTotalNum: 12}
				gomock.InOrder(
					repo.EXPECT().Get(ctx, subscriptionID).Return(bangumi, nil),
					repo.EXPECT().UpdateLastAirEpisode(ctx, subscriptionID, 5).Return(nil),
					parser.EXPECT().GetSeasonEpisodeTotalNum(ctx, 95231, 2, gomock.Any()).Return(12, nil),
				)
				return &Subscriber{repo: repo, metaParser: parser, config: Config{AutoStop: true}}
			},
		},
		{
			name:    "when auto stop disabled and tmdb total grows then updates total and keeps active",
			episode: 12,
			fake: func(t *testing.T) *Subscriber {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				repo := NewMockRepository(ctrl)
				parser := meta.NewMockParser(ctrl)
				bangumi := Bangumi{SubscriptionID: subscriptionID, TMDBID: 95231, Season: 2, EpisodeTotalNum: 12}
				gomock.InOrder(
					repo.EXPECT().Get(ctx, subscriptionID).Return(bangumi, nil),
					repo.EXPECT().UpdateLastAirEpisode(ctx, subscriptionID, 12).Return(nil),
					parser.EXPECT().GetSeasonEpisodeTotalNum(ctx, 95231, 2, gomock.Any()).Return(13, nil),
					repo.EXPECT().UpdateEpisodeTotalNum(ctx, subscriptionID, 13).Return(nil),
				)
				return &Subscriber{repo: repo, metaParser: parser, config: Config{AutoStop: false}}
			},
		},
		{
			name:    "when episode is below local total then refreshes tmdb total",
			episode: 10,
			fake: func(t *testing.T) *Subscriber {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				repo := NewMockRepository(ctrl)
				parser := meta.NewMockParser(ctrl)
				bangumi := Bangumi{SubscriptionID: subscriptionID, TMDBID: 95231, Season: 2, EpisodeTotalNum: 12}
				gomock.InOrder(
					repo.EXPECT().Get(ctx, subscriptionID).Return(bangumi, nil),
					repo.EXPECT().UpdateLastAirEpisode(ctx, subscriptionID, 10).Return(nil),
					parser.EXPECT().GetSeasonEpisodeTotalNum(ctx, 95231, 2, gomock.Any()).Return(13, nil),
					repo.EXPECT().UpdateEpisodeTotalNum(ctx, subscriptionID, 13).Return(nil),
				)
				return &Subscriber{repo: repo, metaParser: parser, config: Config{AutoStop: true}}
			},
		},
		{
			name:    "when auto stop enabled and tmdb total grows then keeps active",
			episode: 12,
			fake: func(t *testing.T) *Subscriber {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				repo := NewMockRepository(ctrl)
				parser := meta.NewMockParser(ctrl)
				bangumi := Bangumi{SubscriptionID: subscriptionID, TMDBID: 95231, Season: 2, EpisodeTotalNum: 12}
				gomock.InOrder(
					repo.EXPECT().Get(ctx, subscriptionID).Return(bangumi, nil),
					repo.EXPECT().UpdateLastAirEpisode(ctx, subscriptionID, 12).Return(nil),
					parser.EXPECT().GetSeasonEpisodeTotalNum(ctx, 95231, 2, gomock.Any()).Return(13, nil),
					repo.EXPECT().UpdateEpisodeTotalNum(ctx, subscriptionID, 13).Return(nil),
				)
				return &Subscriber{repo: repo, metaParser: parser, config: Config{AutoStop: true}}
			},
		},
		{
			name:    "when auto stop enabled and refreshed total is reached then stops subscription",
			episode: 12,
			fake: func(t *testing.T) *Subscriber {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				repo := NewMockRepository(ctrl)
				parser := meta.NewMockParser(ctrl)
				bangumi := Bangumi{SubscriptionID: subscriptionID, TMDBID: 95231, Season: 2, EpisodeTotalNum: 12}
				gomock.InOrder(
					repo.EXPECT().Get(ctx, subscriptionID).Return(bangumi, nil),
					repo.EXPECT().UpdateLastAirEpisode(ctx, subscriptionID, 12).Return(nil),
					parser.EXPECT().GetSeasonEpisodeTotalNum(ctx, 95231, 2, gomock.Any()).Return(12, nil),
					repo.EXPECT().StopSubscription(ctx, subscriptionID).Return(nil),
				)
				return &Subscriber{repo: repo, metaParser: parser, config: Config{AutoStop: true}}
			},
		},
		{
			name:    "when refreshed total decreases and is reached then updates total and stops subscription",
			episode: 12,
			fake: func(t *testing.T) *Subscriber {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				repo := NewMockRepository(ctrl)
				parser := meta.NewMockParser(ctrl)
				bangumi := Bangumi{SubscriptionID: subscriptionID, TMDBID: 95231, Season: 2, EpisodeTotalNum: 13}
				gomock.InOrder(
					repo.EXPECT().Get(ctx, subscriptionID).Return(bangumi, nil),
					repo.EXPECT().UpdateLastAirEpisode(ctx, subscriptionID, 12).Return(nil),
					parser.EXPECT().GetSeasonEpisodeTotalNum(ctx, 95231, 2, gomock.Any()).Return(12, nil),
					repo.EXPECT().UpdateEpisodeTotalNum(ctx, subscriptionID, 12).Return(nil),
					repo.EXPECT().StopSubscription(ctx, subscriptionID).Return(nil),
				)
				return &Subscriber{repo: repo, metaParser: parser, config: Config{AutoStop: true}}
			},
		},
		{
			name:    "when refresh fails then stops by local total",
			episode: 12,
			fake: func(t *testing.T) *Subscriber {
				ctrl := gomock.NewController(t)
				t.Cleanup(ctrl.Finish)

				repo := NewMockRepository(ctrl)
				parser := meta.NewMockParser(ctrl)
				bangumi := Bangumi{SubscriptionID: subscriptionID, TMDBID: 95231, Season: 2, EpisodeTotalNum: 12}
				gomock.InOrder(
					repo.EXPECT().Get(ctx, subscriptionID).Return(bangumi, nil),
					repo.EXPECT().UpdateLastAirEpisode(ctx, subscriptionID, 12).Return(nil),
					parser.EXPECT().GetSeasonEpisodeTotalNum(ctx, 95231, 2, gomock.Any()).Return(0, refreshErr),
					repo.EXPECT().StopSubscription(ctx, subscriptionID).Return(nil),
				)
				return &Subscriber{repo: repo, metaParser: parser, config: Config{AutoStop: true}}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			subscriber := tc.fake(t)

			err := subscriber.HandleEpisodeTransferred(ctx, subscriptionID, tc.episode)

			assert.NoError(t, err)
		})
	}
}
