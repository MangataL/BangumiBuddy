package ass

import (
	"context"

	"github.com/MangataL/BangumiBuddy/pkg/log"
	"github.com/MangataL/BangumiBuddy/pkg/subtitle"
)

type TempFontMetaRepository struct {
	firstRepo  FontMetaRepository
	secondRepo FontMetaRepository
}

func NewTempFontMetaRepository(firstRepo FontMetaRepository, secondRepo FontMetaRepository) FontMetaRepository {
	return &TempFontMetaRepository{firstRepo: firstRepo, secondRepo: secondRepo}
}

func (r *TempFontMetaRepository) Replace(ctx context.Context, fontMetas []FontMeta) error {
	return r.firstRepo.Replace(ctx, fontMetas)
}

func (r *TempFontMetaRepository) Find(ctx context.Context, req FindFontMetaReq) ([]FontMeta, error) {
	fontMetas, err := r.firstRepo.Find(ctx, req)
	if err != nil {
		return nil, err
	}
	fontMetasSecond, err := r.secondRepo.Find(ctx, req)
	if err != nil {
		// 如果第二仓库查询失败，则返回第一仓库的结果
		log.Warnf(ctx, "第二仓库查询失败，返回第一仓库的结果: %v", err)
		return fontMetas, nil
	}
	return append(fontMetas, fontMetasSecond...), nil
}

func (r *TempFontMetaRepository) Stats(ctx context.Context) (subtitle.FontMetaSetStats, error) {
	first, err := r.firstRepo.Stats(ctx)
	if err != nil {
		return first, err
	}
	second, err := r.secondRepo.Stats(ctx)
	second.Total += first.Total
	return second, err
}
