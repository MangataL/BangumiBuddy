package gin

import (
	"github.com/MangataL/BangumiBuddy/internal/auth"
	"github.com/MangataL/BangumiBuddy/internal/downloader"
	"github.com/MangataL/BangumiBuddy/internal/magnet"
	"github.com/MangataL/BangumiBuddy/internal/meta"
	"github.com/MangataL/BangumiBuddy/internal/repository/viper"
	"github.com/MangataL/BangumiBuddy/internal/subscriber"
	"github.com/MangataL/BangumiBuddy/internal/transfer"
	"github.com/MangataL/BangumiBuddy/internal/web"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle"
)

type Dependency struct {
	Authenticator    auth.Authenticator
	TorrentOperator  downloader.TorrentOperator
	ConfigRepo       *viper.Repo
	Subscriber       subscriber.Interface
	Web              web.Interface
	Transfer         transfer.Interface
	Magnet           magnet.Interface
	Parser           meta.Parser
	SubtitleOperator subtitle.Subsetter
}

func New(dep Dependency) *Router {
	return &Router{
		authenticator:     dep.Authenticator,
		torrentOperator:   dep.TorrentOperator,
		repo:              dep.ConfigRepo,
		subscriber:        dep.Subscriber,
		web:               dep.Web,
		transfer:          dep.Transfer,
		magnet:            dep.Magnet,
		metaParser:        dep.Parser,
		subtitleSubsetter: dep.SubtitleOperator,
	}
}

type Router struct {
	authenticator     auth.Authenticator
	torrentOperator   downloader.TorrentOperator
	repo              *viper.Repo
	subscriber        subscriber.Interface
	web               web.Interface
	transfer          transfer.Interface
	magnet            magnet.Interface
	metaParser        meta.Parser
	subtitleSubsetter subtitle.Subsetter
}
