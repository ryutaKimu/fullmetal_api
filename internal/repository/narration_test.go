package repository

import (
	"errors"
	"fullmetal-api/internal/model"
	"testing"
)

func TestNewNarrationRepository(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    int
		wantErr bool
	}{
		{
			name: "正常なJSONを読み込める",
			data: `[
				{"episode":1,"title":{"ja":"鋼の錬金術師","en":null},"narrations":{"ja":["セリフ1"],"en":null}},
				{"episode":2,"title":{"ja":"はじまりの日","en":"The First Day"},"narrations":{"ja":["セリフ2"],"en":["line2"]}}
			]`,
			want: 2,
		},
		{
			name: "空配列でもエラーにならない",
			data: `[]`,
			want: 0,
		},
		{
			name:    "不正なJSONはエラーを返す",
			data:    `[{"episode":1,`,
			wantErr: true,
		},
		{
			name:    "想定と異なる型はエラーを返す",
			data:    `{"episode":1}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewNarrationRepository([]byte(tt.data))

			if tt.wantErr {
				if err == nil {
					t.Fatal("エラーを期待したが nil だった")
				}
				if repo != nil {
					t.Errorf("エラー時の repository は nil であるべき: %+v", repo)
				}
				return
			}

			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if got := len(repo.FindAll()); got != tt.want {
				t.Errorf("件数 = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestNarrationRepository_FindAll(t *testing.T) {
	data := `[
		{"episode":1,"title":{"ja":"鋼の錬金術師","en":null},"narrations":{"ja":["セリフ1","セリフ2"],"en":null}}
	]`

	repo, err := NewNarrationRepository([]byte(data))
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	got := repo.FindAll()
	if len(got) != 1 {
		t.Fatalf("件数 = %d, want 1", len(got))
	}

	n := got[0]
	if n.Episode != 1 {
		t.Errorf("Episode = %d, want 1", n.Episode)
	}
	if n.Title["ja"] == nil || *n.Title["ja"] != "鋼の錬金術師" {
		t.Errorf("Title[ja] = %v, want 鋼の錬金術師", n.Title["ja"])
	}
	// 未翻訳は null のまま保持されることを担保する
	if n.Title["en"] != nil {
		t.Errorf("Title[en] = %v, want nil", *n.Title["en"])
	}
	if n.Narrations["en"] != nil {
		t.Errorf("Narrations[en] = %v, want nil", n.Narrations["en"])
	}
	if len(n.Narrations["ja"]) != 2 {
		t.Errorf("Narrations[ja] の件数 = %d, want 2", len(n.Narrations["ja"]))
	}
}

func TestNarrationRepository_FindEpisode(t *testing.T) {
	data := `[
		{"episode":1,"title":{"ja":"鋼の錬金術師","en":null},"narrations":{"ja":["セリフ1"],"en":null}},
		{"episode":2,"title":{"ja":"はじまりの日","en":"The First Day"},"narrations":{"ja":["セリフ2"],"en":["line2"]}},
		{"episode":3,"title":{"ja":"邪教の街","en":null},"narrations":{"ja":["セリフ3", "セリフ3_1"],"en":null}}
	]`

	tests := []struct {
		name          string
		episode       int
		wantErr       bool
		wantJaTitle   string
		wantEnTitle   *string
		wantJaCount   int
		wantEnPresent bool
	}{
		{
			name:        "先頭のepisodeを取得できる",
			episode:     1,
			wantJaTitle: "鋼の錬金術師",
			wantJaCount: 1,
		},
		{
			name:          "enが存在するepisodeを取得できる",
			episode:       2,
			wantJaTitle:   "はじまりの日",
			wantEnTitle:   new("The First Day"),
			wantJaCount:   1,
			wantEnPresent: true,
		},
		{
			name:        "末尾のepisodeを取得できる",
			episode:     3,
			wantJaTitle: "邪教の街",
			wantJaCount: 2,
		},
		{
			name:    "存在しないepisodeはエラーを返す",
			episode: 99,
			wantErr: true,
		},
		{
			// 末尾の一つ外。比較条件の取り違えを検知する。
			name:    "境界の外側はエラーを返す",
			episode: 4,
			wantErr: true,
		},
		{
			name:    "0はエラーを返す",
			episode: 0,
			wantErr: true,
		},
		{
			name:    "負の数はエラーを返す",
			episode: -1,
			wantErr: true,
		},
	}

	repo, err := NewNarrationRepository([]byte(data))
	if err != nil {
		t.Fatalf("予期しないエラー: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.FindEpisode(tt.episode)

			if tt.wantErr {
				if err == nil {
					t.Fatalf("エラーを期待したが nil だった: %+v", got)
				}
				// 範囲外の episode は不正値も含めてすべて ErrNotFound に寄せる。
				if !errors.Is(err, model.ErrNotFound) {
					t.Errorf("errors.Is(err, model.ErrNotFound) = false, want true (err = %v)", err)
				}
				if got != nil {
					t.Errorf("エラー時の Narration は nil であるべき: %+v", got)
				}
				return
			}

			if err != nil {
				t.Fatalf("予期しないエラー: %v", err)
			}
			if got.Episode != tt.episode {
				t.Errorf("Episode = %d, want %d", got.Episode, tt.episode)
			}
			if got.Title["ja"] == nil || *got.Title["ja"] != tt.wantJaTitle {
				t.Errorf("Title[ja] = %v, want %s", got.Title["ja"], tt.wantJaTitle)
			}

			// 未翻訳は null のまま保持されることを担保する
			switch {
			case tt.wantEnTitle == nil && got.Title["en"] != nil:
				t.Errorf("Title[en] = %v, want nil", *got.Title["en"])
			case tt.wantEnTitle != nil && got.Title["en"] == nil:
				t.Errorf("Title[en] = nil, want %s", *tt.wantEnTitle)
			case tt.wantEnTitle != nil && *got.Title["en"] != *tt.wantEnTitle:
				t.Errorf("Title[en] = %s, want %s", *got.Title["en"], *tt.wantEnTitle)
			}

			if len(got.Narrations["ja"]) != tt.wantJaCount {
				t.Errorf("Narrations[ja] の件数 = %d, want %d", len(got.Narrations["ja"]), tt.wantJaCount)
			}
			if gotEn := got.Narrations["en"] != nil; gotEn != tt.wantEnPresent {
				t.Errorf("Narrations[en] の存在 = %t, want %t", gotEn, tt.wantEnPresent)
			}
		})
	}
}
