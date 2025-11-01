package ass

import (
	"context"
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
			fontMeta.FullName != req.FullName &&
			fontMeta.ChineseFullName != req.FullName {
			continue
		}
		if req.PostScriptName != "" &&
			fontMeta.PostScriptName != req.PostScriptName &&
			fontMeta.ChinesePostScriptName != req.PostScriptName {
			continue
		}
		if req.FamilyName != "" &&
			fontMeta.FamilyName != req.FamilyName &&
			fontMeta.ChineseFamilyName != req.FamilyName {
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
