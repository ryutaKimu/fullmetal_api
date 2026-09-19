package service

import (
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

func ptr(s string) *string {
	return &s
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
				newNarration(1, ptr("鋼の錬金術師"), nil, []string{"セリフ1"}, nil),
				newNarration(2, ptr("はじまりの日"), ptr("The First Day"), []string{"セリフ2"}, []string{"line2"}),
			},
			lang:         "ja",
			wantEpisodes: []int{1, 2},
		},
		{
			name: "enを指定すると英訳済みのデータのみ返る",
			narrations: []model.Narration{
				newNarration(1, ptr("鋼の錬金術師"), nil, []string{"セリフ1"}, nil),
				newNarration(2, ptr("はじまりの日"), ptr("The First Day"), []string{"セリフ2"}, []string{"line2"}),
			},
			lang:         "en",
			wantEpisodes: []int{2},
		},
		{
			name: "titleのみ英訳済みのデータは候補から除外される",
			narrations: []model.Narration{
				newNarration(1, ptr("鋼の錬金術師"), ptr("Fullmetal Alchemist"), []string{"セリフ1"}, nil),
			},
			lang:    "en",
			wantNil: true,
		},
		{
			name: "narrationsのみ英訳済みのデータは候補から除外される",
			narrations: []model.Narration{
				newNarration(1, ptr("鋼の錬金術師"), nil, []string{"セリフ1"}, []string{"line1"}),
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
				newNarration(1, ptr("鋼の錬金術師"), nil, []string{"セリフ1"}, nil),
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
		newNarration(1, ptr("t1"), nil, []string{"n1"}, nil),
		newNarration(2, ptr("t2"), nil, []string{"n2"}, nil),
		newNarration(3, ptr("t3"), nil, []string{"n3"}, nil),
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

func TestNarrationService_GetRandomNarration_呼び出しごとにリポジトリを参照する(t *testing.T) {
	repo := &fakeRepository{
		narrations: []model.Narration{
			newNarration(1, ptr("t1"), nil, []string{"n1"}, nil),
		},
	}
	svc := NewNarrationService(repo)

	svc.GetRandomNarration("ja")
	svc.GetRandomNarration("ja")

	if repo.callCount != 2 {
		t.Errorf("FindAll の呼び出し回数 = %d, want 2", repo.callCount)
	}
}
