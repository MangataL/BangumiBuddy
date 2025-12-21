package ass

import (
	"context"
	"strings"
)

func NewMemoryRepository() FontMetaRepository {
	return &MemoryRepository{fonts: make(map[string]FontMeta)}
}

type MemoryRepository struct {
	fonts map[string]FontMeta
}

// Clean implements FontMetaRepository.
func (r *MemoryRepository) Clean(ctx context.Context) error {
	r.fonts = make(map[string]FontMeta)
	return nil
}

func (r *MemoryRepository) Save(ctx context.Context, fontMetas []FontMeta) error {
	for _, fontMeta := range fontMetas {
		r.fonts[fontMeta.FullName] = fontMeta
	}
	return nil
}

func (r *MemoryRepository) Find(ctx context.Context, req FindFontMetaReq) ([]FontMeta, error) {
	fontMetas := make([]FontMeta, 0)
	for _, fontMeta := range r.fonts {
		if req.FullName != "" &&
			!strings.EqualFold(fontMeta.FullName, req.FullName) &&
			!strings.EqualFold(fontMeta.ChineseFullName, req.FullName) {
			continue
		}
		if req.PostScriptName != "" &&
			!strings.EqualFold(fontMeta.PostScriptName, req.PostScriptName) &&
			!strings.EqualFold(fontMeta.ChinesePostScriptName, req.PostScriptName) {
			continue
		}
		if req.FamilyName != "" &&
			!strings.EqualFold(fontMeta.FamilyName, req.FamilyName) &&
			!strings.EqualFold(fontMeta.ChineseFamilyName, req.FamilyName) {
			continue
		}
		if req.Type != "" && fontMeta.Type != req.Type {
			continue
		}
		fontMetas = append(fontMetas, fontMeta)
	}
	return fontMetas, nil
}

func (r *MemoryRepository) Count(ctx context.Context) (int64, error) {
	return int64(len(r.fonts)), nil
}
