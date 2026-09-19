package repository

import (
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
