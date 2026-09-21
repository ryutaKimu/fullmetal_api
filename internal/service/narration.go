package service

import (
	"fmt"
	"fullmetal-api/internal/model"
	"math/rand/v2"
)

type NarrationRepository interface {
	FindAll() []model.Narration
	FindEpisode(episode int) (*model.Narration, error)
}

type NarrationService struct {
	repo NarrationRepository
}

func NewNarrationService(r NarrationRepository) *NarrationService {
	return &NarrationService{repo: r}
}

// 未対応の言語はマップにキーが存在せずゼロ値のnilが返るため、この判定で弾かれる。
func hasTranslation(n model.Narration, lang string) bool {
	return n.Title[lang] != nil && n.Narrations[lang] != nil
}

func (s NarrationService) GetRandomNarration(lang string) *model.Narration {
	var candidates []model.Narration

	for _, n := range s.repo.FindAll() {
		if hasTranslation(n, lang) {
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

// エピソードが存在しない場合と、存在しても指定言語が未翻訳の場合は、
// どちらもクライアントからは区別できない「見つからない」として扱う。
func (s NarrationService) GetNarrationByEpisode(episode int, lang string) (*model.Narration, error) {
	n, err := s.repo.FindEpisode(episode)
	if err != nil {
		return nil, err
	}

	if !hasTranslation(*n, lang) {
		return nil, fmt.Errorf("episode:%d lang:%s: %w", episode, lang, model.ErrNotFound)
	}

	return n, nil
}
