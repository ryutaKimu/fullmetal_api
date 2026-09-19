package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"fullmetal-api/internal/model"
)

type fakeService struct {
	narration *model.Narration
	gotLang   string
}

func (s *fakeService) GetRandomNarration(lang string) *model.Narration {
	s.gotLang = lang
	return s.narration
}

func ptr(s string) *string {
	return &s
}

func newNarration() *model.Narration {
	return &model.Narration{
		Episode: 1,
		Title:   map[string]*string{"ja": ptr("鋼の錬金術師"), "en": ptr("Fullmetal Alchemist")},
		Narrations: map[string][]string{
			"ja": {"セリフ1", "セリフ2"},
			"en": {"line1", "line2"},
		},
	}
}

func doRequest(t *testing.T, svc NarrationService, target string) *httptest.ResponseRecorder {
	t.Helper()

	h := NewNarrationHandler(svc)
	req := httptest.NewRequest(http.MethodGet, target, nil)
	rec := httptest.NewRecorder()
	h.NarrationHandler(rec, req)

	return rec
}

func TestNarrationHandler_正常系(t *testing.T) {
	tests := []struct {
		name           string
		target         string
		wantLang       string
		wantTitle      string
		wantNarrations []string
	}{
		{
			name:           "langを省略すると日本語で返る",
			target:         "/narrations/random",
			wantLang:       "ja",
			wantTitle:      "鋼の錬金術師",
			wantNarrations: []string{"セリフ1", "セリフ2"},
		},
		{
			name:           "lang=jaを指定すると日本語で返る",
			target:         "/narrations/random?lang=ja",
			wantLang:       "ja",
			wantTitle:      "鋼の錬金術師",
			wantNarrations: []string{"セリフ1", "セリフ2"},
		},
		{
			name:           "lang=enを指定すると英語で返る",
			target:         "/narrations/random?lang=en",
			wantLang:       "en",
			wantTitle:      "Fullmetal Alchemist",
			wantNarrations: []string{"line1", "line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeService{narration: newNarration()}
			rec := doRequest(t, svc, tt.target)

			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			if svc.gotLang != tt.wantLang {
				t.Errorf("サービスに渡された lang = %q, want %q", svc.gotLang, tt.wantLang)
			}

			var res response
			if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
				t.Fatalf("レスポンスのデコードに失敗: %v (body=%s)", err, rec.Body.String())
			}
			if res.Error != "" {
				t.Errorf("error = %q, want 空文字", res.Error)
			}
			if res.Data == nil {
				t.Fatal("data を期待したが null だった")
			}
			if res.Data.Episode != 1 {
				t.Errorf("episode = %d, want 1", res.Data.Episode)
			}
			if res.Data.Title != tt.wantTitle {
				t.Errorf("title = %q, want %q", res.Data.Title, tt.wantTitle)
			}
			if !slices.Equal(res.Data.Narrations, tt.wantNarrations) {
				t.Errorf("narrations = %v, want %v", res.Data.Narrations, tt.wantNarrations)
			}
		})
	}
}

func TestNarrationHandler_未対応の言語は400を返す(t *testing.T) {
	for _, target := range []string{
		"/narrations/random?lang=fr",
		"/narrations/random?lang=JA",
		"/narrations/random?lang=%20",
	} {
		t.Run(target, func(t *testing.T) {
			svc := &fakeService{narration: newNarration()}
			rec := doRequest(t, svc, target)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
			if got := rec.Header().Get("Content-Type"); got != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", got)
			}
			if svc.gotLang != "" {
				t.Errorf("サービスは呼ばれないはずだが lang=%q で呼ばれた", svc.gotLang)
			}

			var res response
			if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
				t.Fatalf("レスポンスのデコードに失敗: %v (body=%s)", err, rec.Body.String())
			}
			// data はエラー時も常に null で返す（利用側が data の有無だけで判定できるようにするため）
			if res.Data != nil {
				t.Errorf("data = %+v, want null", res.Data)
			}
			if res.Error != "the language is not supported" {
				t.Errorf("error = %q, want %q", res.Error, "the language is not supported")
			}
		})
	}
}

func TestNarrationHandler_該当データがない場合は200でdataがnull(t *testing.T) {
	svc := &fakeService{narration: nil}
	rec := doRequest(t, svc, "/narrations/random?lang=en")

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var raw map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("レスポンスのデコードに失敗: %v (body=%s)", err, rec.Body.String())
	}
	// data キー自体は省略せず、null として必ず含める
	v, ok := raw["data"]
	if !ok {
		t.Fatalf("data キーが存在しない (body=%s)", rec.Body.String())
	}
	if v != nil {
		t.Errorf("data = %v, want null", v)
	}
	if _, ok := raw["error"]; ok {
		t.Errorf("error キーは含めない想定 (body=%s)", rec.Body.String())
	}
}
