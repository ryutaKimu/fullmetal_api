package data

import (
	"encoding/json"
	"testing"

	"fullmetal-api/internal/model"
)

// 埋め込みJSONがデータ規約（docs/faq.md）を満たしているかを検証する。
// 未翻訳は null のみで表し、空文字・空配列は使わない。
func TestNarrationsJSON(t *testing.T) {
	var narrations []model.Narration
	if err := json.Unmarshal(NarrationsJSON, &narrations); err != nil {
		t.Fatalf("narrations.json のパースに失敗: %v", err)
	}

	if len(narrations) == 0 {
		t.Fatal("narrations.json が空")
	}

	langs := []string{"ja", "en"}
	seen := map[int]bool{}

	for _, n := range narrations {
		if n.Episode <= 0 {
			t.Errorf("episode = %d, want 1以上", n.Episode)
		}
		// episode で一意に特定できることを担保する
		if seen[n.Episode] {
			t.Errorf("episode %d が重複している", n.Episode)
		}
		seen[n.Episode] = true

		for _, lang := range langs {
			title, ok := n.Title[lang]
			if !ok {
				t.Errorf("episode %d: title に %q キーがない", n.Episode, lang)
			}
			if title != nil && *title == "" {
				t.Errorf("episode %d: title[%s] が空文字。未翻訳は null で表す", n.Episode, lang)
			}

			lines, ok := n.Narrations[lang]
			if !ok {
				t.Errorf("episode %d: narrations に %q キーがない", n.Episode, lang)
			}
			if lines != nil && len(lines) == 0 {
				t.Errorf("episode %d: narrations[%s] が空配列。未翻訳は null で表す", n.Episode, lang)
			}
			for i, line := range lines {
				if line == "" {
					t.Errorf("episode %d: narrations[%s][%d] が空文字", n.Episode, lang, i)
				}
			}
		}

		// 日本語は必須。これが欠けるとAPIが1件も返せなくなる
		if n.Title["ja"] == nil || n.Narrations["ja"] == nil {
			t.Errorf("episode %d: 日本語のタイトル・ナレーションは必須", n.Episode)
		}
	}
}
