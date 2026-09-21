package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"fullmetal-api/internal/model"
)

type fakeService struct {
	narration  *model.Narration
	err        error
	gotLang    string
	gotEpisode int
}

func (s *fakeService) GetRandomNarration(lang string) *model.Narration {
	s.gotLang = lang
	return s.narration
}

func (s *fakeService) GetNarrationByEpisode(episode int, lang string) (*model.Narration, error) {
	s.gotLang = lang
	s.gotEpisode = episode
	if s.err != nil {
		return nil, s.err
	}
	return s.narration, nil
}

func newNarration() *model.Narration {
	return &model.Narration{
		Episode: 1,
		Title:   map[string]*string{"ja": new("鋼の錬金術師"), "en": new("Fullmetal Alchemist")},
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

func doEpisodeRequest(t *testing.T, svc NarrationService, episode, query string) *httptest.ResponseRecorder {
	t.Helper()

	h := NewNarrationHandler(svc)
	req := httptest.NewRequest(http.MethodGet, "/narrations/"+episode+query, nil)
	// muxを経由せずハンドラを直接呼ぶため、パスパラメーターは手で設定する
	req.SetPathValue("episode", episode)
	rec := httptest.NewRecorder()
	h.OneNarrationHandler(rec, req)

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
			// charset が無いとブラウザで日本語が文字化けするため、値まで含めて検証する
			if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q, want application/json; charset=utf-8", got)
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
			// charset が無いとブラウザで日本語が文字化けするため、値まで含めて検証する
			if got := rec.Header().Get("Content-Type"); got != "application/json; charset=utf-8" {
				t.Errorf("Content-Type = %q, want application/json; charset=utf-8", got)
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

// serviceが非nilのtitleに絞り込む前提が崩れても、デリファレンスでpanicせず500を返すことを固定する。
func TestNarrationHandler_titleがnilなら500を返す(t *testing.T) {
	svc := &fakeService{narration: &model.Narration{
		Episode:    1,
		Title:      map[string]*string{"en": nil},
		Narrations: map[string][]string{"en": {"line1", "line2"}},
	}}
	rec := doRequest(t, svc, "/narrations/random?lang=en")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
	// charset が無いとブラウザで日本語が文字化けするため、値まで含めて検証する
	if got := rec.Header().Get("Content-Type"); got != contentTypeJSON {
		t.Errorf("Content-Type = %q, want %q", got, contentTypeJSON)
	}

	var res response
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("レスポンスのデコードに失敗: %v (body=%s)", err, rec.Body.String())
	}
	// data はエラー時も常に null で返す（利用側が data の有無だけで判定できるようにするため）
	if res.Data != nil {
		t.Errorf("data = %+v, want null", res.Data)
	}
	if res.Error != "internal server error" {
		t.Errorf("error = %q, want %q", res.Error, "internal server error")
	}
}

func TestOneNarrationHandler_正常系(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		wantLang       string
		wantTitle      string
		wantNarrations []string
	}{
		{
			name:           "langを省略すると日本語で返る",
			query:          "",
			wantLang:       "ja",
			wantTitle:      "鋼の錬金術師",
			wantNarrations: []string{"セリフ1", "セリフ2"},
		},
		{
			name:           "lang=enを指定すると英語で返る",
			query:          "?lang=en",
			wantLang:       "en",
			wantTitle:      "Fullmetal Alchemist",
			wantNarrations: []string{"line1", "line2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &fakeService{narration: newNarration()}
			rec := doEpisodeRequest(t, svc, "5", tt.query)

			if rec.Code != http.StatusOK {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
			}
			if got := rec.Header().Get("Content-Type"); got != contentTypeJSON {
				t.Errorf("Content-Type = %q, want %q", got, contentTypeJSON)
			}
			// パスパラメーターがintへ変換されてサービスに渡ることを確認する
			if svc.gotEpisode != 5 {
				t.Errorf("サービスに渡された episode = %d, want 5", svc.gotEpisode)
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
			if res.Data.Title != tt.wantTitle {
				t.Errorf("title = %q, want %q", res.Data.Title, tt.wantTitle)
			}
			if !slices.Equal(res.Data.Narrations, tt.wantNarrations) {
				t.Errorf("narrations = %v, want %v", res.Data.Narrations, tt.wantNarrations)
			}
		})
	}
}

func TestOneNarrationHandler_数値でないepisodeは400を返す(t *testing.T) {
	for _, episode := range []string{"abc", "1.5", "5x", ""} {
		t.Run(episode, func(t *testing.T) {
			svc := &fakeService{narration: newNarration()}
			rec := doEpisodeRequest(t, svc, episode, "")

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
			if svc.gotEpisode != 0 {
				t.Errorf("サービスは呼ばれないはずだが episode=%d で呼ばれた", svc.gotEpisode)
			}

			var res response
			if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
				t.Fatalf("レスポンスのデコードに失敗: %v (body=%s)", err, rec.Body.String())
			}
			if res.Data != nil {
				t.Errorf("data = %+v, want null", res.Data)
			}
			if res.Error != "episode must be an integer" {
				t.Errorf("error = %q, want %q", res.Error, "episode must be an integer")
			}
		})
	}
}

func TestOneNarrationHandler_未対応の言語は400を返す(t *testing.T) {
	svc := &fakeService{narration: newNarration()}
	rec := doEpisodeRequest(t, svc, "5", "?lang=fr")

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if svc.gotLang != "" {
		t.Errorf("サービスは呼ばれないはずだが lang=%q で呼ばれた", svc.gotLang)
	}

	var res response
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("レスポンスのデコードに失敗: %v (body=%s)", err, rec.Body.String())
	}
	if res.Error != "the language is not supported" {
		t.Errorf("error = %q, want %q", res.Error, "the language is not supported")
	}
}

// エピソード不在と未翻訳はserviceが同じErrNotFoundで返すため、どちらも404になる。
func TestOneNarrationHandler_該当データがない場合は404を返す(t *testing.T) {
	svc := &fakeService{err: fmt.Errorf("episode:999: %w", model.ErrNotFound)}
	rec := doEpisodeRequest(t, svc, "999", "")

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
	if got := rec.Header().Get("Content-Type"); got != contentTypeJSON {
		t.Errorf("Content-Type = %q, want %q", got, contentTypeJSON)
	}

	var res response
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("レスポンスのデコードに失敗: %v (body=%s)", err, rec.Body.String())
	}
	if res.Data != nil {
		t.Errorf("data = %+v, want null", res.Data)
	}
	if res.Error != "narration not found" {
		t.Errorf("error = %q, want %q", res.Error, "narration not found")
	}
}

// ErrNotFound以外のエラーを404へ吸い込ませず、500として表面化させることを固定する。
func TestOneNarrationHandler_想定外のエラーは500を返す(t *testing.T) {
	svc := &fakeService{err: errors.New("unexpected failure")}
	rec := doEpisodeRequest(t, svc, "5", "")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var res response
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("レスポンスのデコードに失敗: %v (body=%s)", err, rec.Body.String())
	}
	if res.Data != nil {
		t.Errorf("data = %+v, want null", res.Data)
	}
	if res.Error != "internal server error" {
		t.Errorf("error = %q, want %q", res.Error, "internal server error")
	}
}

func TestOneNarrationHandler_titleがnilなら500を返す(t *testing.T) {
	svc := &fakeService{narration: &model.Narration{
		Episode:    5,
		Title:      map[string]*string{"en": nil},
		Narrations: map[string][]string{"en": {"line1", "line2"}},
	}}
	rec := doEpisodeRequest(t, svc, "5", "?lang=en")

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var res response
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("レスポンスのデコードに失敗: %v (body=%s)", err, rec.Body.String())
	}
	if res.Error != "internal server error" {
		t.Errorf("error = %q, want %q", res.Error, "internal server error")
	}
}
