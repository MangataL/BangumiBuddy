package discovery

import "context"

//go:generate mockgen -destination interface_mock.go -source $GOFILE -package $GOPACKAGE

type Interface interface {
	ListBangumis(ctx context.Context, req ListBangumiReq) ([]BangumiCandidate, error)
	Search(ctx context.Context, req SearchReq) (SearchResp, error)
	BatchReleaseGroups(ctx context.Context, req BatchReleaseGroupsReq) (BatchReleaseGroupsResp, error)
	GetBangumiDetail(ctx context.Context, id string) (BangumiDetail, error)
	BuildCandidateRSS(req CandidateRSSReq) (string, error)
}
