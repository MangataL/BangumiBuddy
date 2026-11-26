package anito

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/nssteinbrenner/anitogo"
)

type ReleaseKind int

const (
	UnknownRelease ReleaseKind = iota
	SingleEpisode
	EpisodeCollection
)

var (
	collectionMarker = regexp.MustCompile(`合集|全集|合輯|合辑|全\s*\d+\s*[话話集]|(?:\[|【|第)\s*(?:第\s*)?\d{1,3}\s*[话話集]?\s*[-～~–—至到]\s*(?:第\s*)?\d{1,3}\s*[话話集\]】]`)
	extraMarker      = regexp.MustCompile(`特别篇|特別篇|总集篇|總集篇|剧场版|劇場版|(?i:\b(?:sp|ova|oad|pv|ncop|nced)\s*\d+\b)`)
)

func ClassifyRelease(title string) ReleaseKind {
	elements := anitogo.Parse(normalizeFilenameForAnitogo(title), anitogo.DefaultOptions)
	for _, kind := range elements.AnimeType {
		if !strings.EqualFold(kind, "TV") {
			return UnknownRelease
		}
	}
	if extraMarker.MatchString(title) {
		return UnknownRelease
	}
	if len(elements.EpisodeNumber) > 1 || collectionMarker.MatchString(title) {
		return EpisodeCollection
	}
	for _, info := range elements.ReleaseInformation {
		if strings.EqualFold(info, "Batch") || strings.EqualFold(info, "Complete") {
			return EpisodeCollection
		}
	}
	if len(elements.EpisodeNumber) == 1 {
		if n, err := strconv.Atoi(elements.EpisodeNumber[0]); err == nil && n > 0 {
			return SingleEpisode
		}
	}
	return UnknownRelease
}
