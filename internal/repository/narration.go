package repository

import (
	"encoding/json"
	"fmt"
	"fullmetal-api/internal/model"
)

type NarrationRepository struct {
	narration []model.Narration
}

func NewNarrationRepository(data []byte) (*NarrationRepository, error) {
	var narrations []model.Narration
	if err := json.Unmarshal(data, &narrations); err != nil {
		return nil, err
	}
	return &NarrationRepository{narration: narrations}, nil
}

func (r *NarrationRepository) FindAll() []model.Narration {
	return r.narration
}

func (r *NarrationRepository) FindEpisode(episode int) (*model.Narration, error) {
	for _, n := range r.narration {
		if n.Episode == episode {
			return &n, nil
		}
	}
	return nil, fmt.Errorf("episode:%d: %w", episode, model.ErrNotFound)
}
