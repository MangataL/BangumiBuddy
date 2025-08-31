package ass

import "context"

type TempFontMetaRepository struct {
	firstRepo  FontMetaRepository
	secondRepo FontMetaRepository
}

func NewTempFontMetaRepository(firstRepo FontMetaRepository, secondRepo FontMetaRepository) FontMetaRepository {
	return &TempFontMetaRepository{firstRepo: firstRepo, secondRepo: secondRepo}
}

func (r *TempFontMetaRepository) Save(ctx context.Context, fontMetas []FontMeta) error {
	return r.firstRepo.Save(ctx, fontMetas)
}

func (r *TempFontMetaRepository) Find(ctx context.Context, req FindFontMetaReq) ([]FontMeta, error) {
	fontMetas, err := r.firstRepo.Find(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(fontMetas) != 0 {
		return fontMetas, nil
	}
	return r.secondRepo.Find(ctx, req)
}

func (r *TempFontMetaRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.firstRepo.Count(ctx)
	if err != nil {
		return 0, err
	}
	secondCount, err := r.secondRepo.Count(ctx)
	if err != nil {
		return 0, err
	}
	return count + secondCount, nil
}

func (r *TempFontMetaRepository) Clean(ctx context.Context) error {
	return r.firstRepo.Clean(ctx)
}