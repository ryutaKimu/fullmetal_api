package service

import (
	"errors"
	"fmt"
	"slices"
	"testing"

	"fullmetal-api/internal/model"
)

type fakeRepository struct {
	narrations []model.Narration
	callCount  int
}

func (r *fakeRepository) FindAll() []model.Narration {
	r.callCount++
	return r.narrations
}

func (r *fakeRepository) FindEpisode(episode int) (*model.Narration, error) {
	for i := range r.narrations {
		if r.narrations[i].Episode == episode {
			return &r.narrations[i], nil
		}
	}
	// 本物のリポジトリと同じく model.ErrNotFound を包む。包まないと errors.Is での判定が
	// テストだけ通らない／通ってしまう状態になり、fakeが本番の振る舞いから乖離する。
	return nil, fmt.Errorf("episode:%d: %w", episode, model.ErrNotFound)
}

func newNarration(episode int, jaTitle, enTitle *string, ja, en []string) model.Narration {
	return model.Narration{
		Episode:    episode,
		Title:      map[string]*string{"ja": jaTitle, "en": enTitle},
		Narrations: map[string][]string{"ja": ja, "en": en},
	}
}

func TestNarrationService_GetRandomNarration(t *testing.T) {
	tests := []struct {
		name         string
		narrations   []model.Narration
		lang         string
		wantNil      bool
		wantEpisodes []int // 返りうる episode の集合
	}{
		{
			name: "jaを指定すると日本語が揃ったデータのみ返る",
			narrations: []model.Narration{
				newNarration(1, new("鋼の錬金術師"), nil, []string{"セリフ1"}, nil),
				newNarration(2, new("はじまりの日"), new("The First Day"), []string{"セリフ2"}, []string{"line2"}),
			},
			lang:         "ja",
			wantEpisodes: []int{1, 2},
		},
		{
			name: "enを指定すると英訳済みのデータのみ返る",
			narrations: []model.Narration{
				newNarration(1, new("鋼の錬金術師"), nil, []string{"セリフ1"}, nil),
				newNarration(2, new("はじまりの日"), new("The First Day"), []string{"セリフ2"}, []string{"line2"}),
			},
			lang:         "en",
			wantEpisodes: []int{2},
		},
		{
			name: "titleのみ英訳済みのデータは候補から除外される",
			narrations: []model.Narration{
				newNarration(1, new("鋼の錬金術師"), new("Fullmetal Alchemist"), []string{"セリフ1"}, nil),
			},
			lang:    "en",
			wantNil: true,
		},
		{
			name: "narrationsのみ英訳済みのデータは候補から除外される",
			narrations: []model.Narration{
				newNarration(1, new("鋼の錬金術師"), nil, []string{"セリフ1"}, []string{"line1"}),
			},
			lang:    "en",
			wantNil: true,
		},
		{
			name:       "データが0件ならnilを返す",
			narrations: nil,
			lang:       "ja",
			wantNil:    true,
		},
		{
			name: "未対応の言語はnilを返す",
			narrations: []model.Narration{
				newNarration(1, new("鋼の錬金術師"), nil, []string{"セリフ1"}, nil),
			},
			lang:    "fr",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewNarrationService(&fakeRepository{narrations: tt.narrations})

			// ランダム選択のため、複数回試行して候補外が混ざらないことを確認する
			for range 50 {
				got := svc.GetRandomNarration(tt.lang)

				if tt.wantNil {
					if got != nil {
						t.Fatalf("nil を期待したが episode %d が返った", got.Episode)
					}
					continue
				}

				if got == nil {
					t.Fatal("ナレーションを期待したが nil だった")
				}
				if got.Title[tt.lang] == nil || got.Narrations[tt.lang] == nil {
					t.Fatalf("episode %d は %s が未翻訳のため返ってはいけない", got.Episode, tt.lang)
				}
				if !slices.Contains(tt.wantEpisodes, got.Episode) {
					t.Fatalf("episode = %d, want いずれか %v", got.Episode, tt.wantEpisodes)
				}
			}
		})
	}
}

func TestNarrationService_GetRandomNarration_候補が複数なら偏らず選ばれる(t *testing.T) {
	narrations := []model.Narration{
		newNarration(1, new("t1"), nil, []string{"n1"}, nil),
		newNarration(2, new("t2"), nil, []string{"n2"}, nil),
		newNarration(3, new("t3"), nil, []string{"n3"}, nil),
	}
	svc := NewNarrationService(&fakeRepository{narrations: narrations})

	seen := map[int]bool{}
	for range 300 {
		got := svc.GetRandomNarration("ja")
		if got == nil {
			t.Fatal("ナレーションを期待したが nil だった")
		}
		seen[got.Episode] = true
	}

	// 300回試行して1件も選ばれない episode があれば、ランダム選択が機能していない
	for _, n := range narrations {
		if !seen[n.Episode] {
			t.Errorf("episode %d が一度も選ばれなかった", n.Episode)
		}
	}
}

func TestNarrationService_GetNarrationByEpisode(t *testing.T) {
	narrations := []model.Narration{
		newNarration(1, new("第1話"), new("Episode 1"), []string{"セリフ1"}, []string{"line1"}),
		newNarration(2, new("第2話"), new("Episode 2"), []string{"セリフ2"}, nil),
		newNarration(3, new("第3話"), nil, []string{"セリフ3"}, []string{"line3"}),
		newNarration(4, new("第4話"), nil, []string{"セリフ4"}, nil),
	}

	tests := []struct {
		name        string
		narrations  []model.Narration
		episode     int
		lang        string
		wantEpisode int
		wantErr     bool
	}{
		{
			name:        "jaが揃っていれば指定エピソードが返る",
			narrations:  narrations,
			episode:     1,
			lang:        "ja",
			wantEpisode: 1,
		},
		{
			name:        "enが揃っていれば指定エピソードが返る",
			narrations:  narrations,
			episode:     1,
			lang:        "en",
			wantEpisode: 1,
		},
		{
			name:       "narrationsのみ未翻訳ならエラー",
			narrations: narrations,
			episode:    2,
			lang:       "en",
			wantErr:    true,
		},
		{
			name:       "titleのみ未翻訳ならエラー",
			narrations: narrations,
			episode:    3,
			lang:       "en",
			wantErr:    true,
		},
		{
			name:       "titleとnarrationsの両方が未翻訳ならエラー",
			narrations: narrations,
			episode:    4,
			lang:       "en",
			wantErr:    true,
		},
		{
			name:       "存在しないエピソードはエラー",
			narrations: narrations,
			episode:    999,
			lang:       "ja",
			wantErr:    true,
		},
		{
			name:       "データが0件ならエラー",
			narrations: nil,
			episode:    1,
			lang:       "ja",
			wantErr:    true,
		},
		{
			name:       "未対応の言語はエラー",
			narrations: narrations,
			episode:    1,
			lang:       "fr",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewNarrationService(&fakeRepository{narrations: tt.narrations})

			got, err := svc.GetNarrationByEpisode(tt.episode, tt.lang)

			if tt.wantErr {
				// エラーメッセージの文言に依存させないため、番兵エラーで判定する
				if !errors.Is(err, model.ErrNotFound) {
					t.Fatalf("err = %v, want model.ErrNotFound", err)
				}
				if got != nil {
					t.Errorf("エラー時の戻り値 = episode %d, want nil", got.Episode)
				}
				return
			}

			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if got == nil {
				t.Fatal("ナレーションを期待したが nil だった")
			}
			if got.Episode != tt.wantEpisode {
				t.Errorf("episode = %d, want %d", got.Episode, tt.wantEpisode)
			}
			if got.Title[tt.lang] == nil || got.Narrations[tt.lang] == nil {
				t.Errorf("episode %d は %s が未翻訳のため返ってはいけない", got.Episode, tt.lang)
			}
		})
	}
}

func TestNarrationService_GetRandomNarration_呼び出しごとにリポジトリを参照する(t *testing.T) {
	repo := &fakeRepository{
		narrations: []model.Narration{
			newNarration(1, new("t1"), nil, []string{"n1"}, nil),
		},
	}
	svc := NewNarrationService(repo)

	svc.GetRandomNarration("ja")
	svc.GetRandomNarration("ja")

	if repo.callCount != 2 {
		t.Errorf("FindAll の呼び出し回数 = %d, want 2", repo.callCount)
	}
}
