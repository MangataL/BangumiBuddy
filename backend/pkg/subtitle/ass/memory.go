package ass

import (
	"context"
	"strings"
	"sync"

	"github.com/MangataL/BangumiBuddy/pkg/subtitle"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle/ass/freetype"
)

func NewMemoryRepository() FontMetaRepository {
	return &MemoryRepository{fonts: make(map[string]FontMeta)}
}

type MemoryRepository struct {
	mu    sync.RWMutex
	fonts map[string]FontMeta
}

func (r *MemoryRepository) Replace(ctx context.Context, fontMetas []FontMeta) error {
	fonts := make(map[string]FontMeta, len(fontMetas))
	for _, m := range fontMetas {
		fonts[strings.ToLower(m.FullName)] = m
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fonts = fonts
	return nil
}

func hasFontName(m FontMeta, kind uint16, name string) bool {
	key := FontNameKey(name)
	for _, alias := range m.Aliases {
		if alias.Kind == kind && FontNameKey(alias.Value) == key {
			return true
		}
	}
	return false
}

func (r *MemoryRepository) Find(ctx context.Context, req FindFontMetaReq) ([]FontMeta, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fontMetas := make([]FontMeta, 0)
	for _, fontMeta := range r.fonts {
		if req.FullName != "" &&
			!strings.EqualFold(fontMeta.FullName, req.FullName) &&
			!strings.EqualFold(fontMeta.ChineseFullName, req.FullName) && !hasFontName(fontMeta, freetype.NameIDFullName, req.FullName) {
			continue
		}
		if req.PostScriptName != "" &&
			!strings.EqualFold(fontMeta.PostScriptName, req.PostScriptName) &&
			!strings.EqualFold(fontMeta.ChinesePostScriptName, req.PostScriptName) && !hasFontName(fontMeta, freetype.NameIDPostScript, req.PostScriptName) {
			continue
		}
		if req.FamilyName != "" &&
			!strings.EqualFold(fontMeta.FamilyName, req.FamilyName) &&
			!strings.EqualFold(fontMeta.ChineseFamilyName, req.FamilyName) && !hasFontName(fontMeta, freetype.NameIDFamily, req.FamilyName) {
			continue
		}
		if req.Type != "" && fontMeta.Type != req.Type {
			continue
		}
		fontMetas = append(fontMetas, fontMeta)
	}
	return fontMetas, nil
}

func (r *MemoryRepository) Stats(ctx context.Context) (subtitle.FontMetaSetStats, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return subtitle.FontMetaSetStats{Total: int64(len(r.fonts))}, nil
}
