package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"fullmetal-api/internal/handler"
	"fullmetal-api/internal/model"
)

type fakeService struct {
	gotEpisode int
}

func (s *fakeService) GetRandomNarration(lang string) *model.Narration {
	return newNarration()
}

func (s *fakeService) GetNarrationByEpisode(episode int, lang string) (*model.Narration, error) {
	s.gotEpisode = episode
	return newNarration(), nil
}

func newNarration() *model.Narration {
	title := "鋼の錬金術師"
	return &model.Narration{
		Episode:    1,
		Title:      map[string]*string{"ja": &title},
		Narrations: map[string][]string{"ja": {"セリフ1"}},
	}
}

func doRequest(t *testing.T, target string) (*fakeService, *httptest.ResponseRecorder) {
	t.Helper()

	svc := &fakeService{}
	mux := New(handler.NewNarrationHandler(svc))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, target, nil))

	return svc, rec
}

func TestV1ルートがハンドラに到達する(t *testing.T) {
	t.Run("random", func(t *testing.T) {
		_, rec := doRequest(t, "/v1/narrations/random")
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
	})

	t.Run("episode指定", func(t *testing.T) {
		svc, rec := doRequest(t, "/v1/narrations/5")
		if rec.Code != http.StatusOK {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
		}
		if svc.gotEpisode != 5 {
			t.Errorf("episode = %d, want 5", svc.gotEpisode)
		}
	})
}

func TestLegacyパスはv1へリダイレクトする(t *testing.T) {
	tests := []struct {
		target       string
		wantLocation string
	}{
		{"/narrations/random", "/v1/narrations/random"},
		{"/narrations/5", "/v1/narrations/5"},
		{"/narrations/random?lang=en", "/v1/narrations/random?lang=en"},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			_, rec := doRequest(t, tt.target)
			if rec.Code != http.StatusPermanentRedirect {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusPermanentRedirect)
			}
			if got := rec.Header().Get("Location"); got != tt.wantLocation {
				t.Errorf("Location = %q, want %q", got, tt.wantLocation)
			}
		})
	}
}

func Test未定義のパスは404を返す(t *testing.T) {
	for _, target := range []string{"/", "/v1", "/v1/narrations", "/v2/narrations/random"} {
		t.Run(target, func(t *testing.T) {
			_, rec := doRequest(t, target)
			if rec.Code != http.StatusNotFound {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
			}
		})
	}
}
