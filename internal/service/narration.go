package service

import (
	"fullmetal-api/internal/model"
	"math/rand/v2"
)

type NarrationRepository interface {
	FindAll() []model.Narration
}

type NarrationService struct {
	repo NarrationRepository
}

func NewNarrationService(r NarrationRepository) *NarrationService {
	return &NarrationService{repo: r}
}

func (s NarrationService) GetRandomNarration(lang string) *model.Narration {
	var candidates []model.Narration

	for _, n := range s.repo.FindAll() {
		// 指定言語のタイトルとナレーションが両方存在するデータのみ候補とする。
		if n.Title[lang] != nil && n.Narrations[lang] != nil {
			candidates = append(candidates, n)
		}
	}

	// rand.IntNの引数値が0だとpanicでプログラムが異常終了するため、早期リターンで事前に防止。
	if len(candidates) == 0 {
		return nil
	}
	picked := candidates[rand.IntN(len(candidates))]

	return &picked
}
