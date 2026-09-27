package handler

import (
	"encoding/json"
	"errors"
	"fullmetal-api/internal/model"
	"log"
	"net/http"
	"strconv"
)

// charsetを明示しないと、ブラウザによっては文字コードを誤推測して日本語が文字化けする。
const contentTypeJSON = "application/json; charset=utf-8"

type narrationResponse struct {
	Episode    int      `json:"episode"`
	Title      string   `json:"title"`
	Narrations []string `json:"narrations"`
}

type response struct {
	Data  *narrationResponse `json:"data"`
	Error string             `json:"error,omitempty"`
}

type responses struct {
	Data  []*narrationResponse `json:"data"`
	Error string               `json:"error,omitempty"`
}

type NarrationService interface {
	GetRandomNarration(lang string) *model.Narration
	GetRandomNarrations(episodes int, lang string) ([]*model.Narration, error)
	GetNarrationByEpisode(episode int, lang string) (*model.Narration, error)
}

type NarrationHandler struct {
	service NarrationService
}

func NewNarrationHandler(service NarrationService) *NarrationHandler {
	return &NarrationHandler{service: service}
}

// デフォルトでは日本語に設定することを担保
func resolveLang(r *http.Request) (string, bool) {
	lang := r.URL.Query().Get("lang")
	if lang == "" {
		lang = "ja"
	}

	// Todo: 言語リストを作成し、パラメーターがリストに含まれるかの判定を行うようにする。
	if lang != "ja" && lang != "en" {
		return "", false
	}

	return lang, true
}

func writeNarration(w http.ResponseWriter, narration *model.Narration, lang string) {
	title := narration.Title[lang]
	if title == nil {
		// serviceで非nilに絞り込み済みなので、ここに来るのは不変条件の破れ
		log.Printf("unexpected nil title: episode=%d lang=%s", narration.Episode, lang)
		jsonEncoder(w, http.StatusInternalServerError, response{Data: nil, Error: "internal server error"})
		return
	}

	res := narrationResponse{
		Episode:    narration.Episode,
		Title:      *title,
		Narrations: narration.Narrations[lang],
	}
	jsonEncoder(w, http.StatusOK, response{Data: &res})
}

func writeNarrations(w http.ResponseWriter, narrations []*model.Narration, lang string) {
	res := make([]*narrationResponse, 0, len(narrations))
	for _, n := range narrations {
		title := n.Title[lang]
		if title == nil {
			log.Printf("unexpected nil title: episode=%d lang=%s", n.Episode, lang)
			jsonEncoder(w, http.StatusInternalServerError, response{Data: nil, Error: "internal server error"})
			return
		}
		narration := narrationResponse{
			Episode:    n.Episode,
			Title:      *title,
			Narrations: n.Narrations[lang],
		}
		res = append(res, &narration)
	}
	jsonEncoder(w, http.StatusOK, responses{Data: res, Error: ""})
}

func (h *NarrationHandler) NarrationHandler(w http.ResponseWriter, r *http.Request) {
	lang, ok := resolveLang(r)
	if !ok {
		jsonEncoder(w, http.StatusBadRequest, response{Data: nil, Error: "the language is not supported"})
		return
	}

	narration := h.service.GetRandomNarration(lang)
	if narration == nil {
		jsonEncoder(w, http.StatusOK, response{})
		return
	}

	writeNarration(w, narration, lang)
}

func (h *NarrationHandler) RandomNarrationsHandler(w http.ResponseWriter, r *http.Request) {
	lang, ok := resolveLang(r)
	if !ok {
		jsonEncoder(w, http.StatusBadRequest, response{Data: nil, Error: "the language is not supported"})
		return
	}

	episodes, err := strconv.Atoi(r.PathValue("episodes"))

	if err != nil {
		jsonEncoder(w, http.StatusBadRequest, response{Data: nil, Error: "episode must be an integer"})
		return
	}
	narrations, err := h.service.GetRandomNarrations(episodes, lang)
	if err != nil {
		// 範囲外はリクエスト不正、候補不足は指定件数分のナレーションが見つからないものとして扱う。
		if errors.Is(err, model.ErrInvalidEpisodes) {
			jsonEncoder(w, http.StatusBadRequest, response{Data: nil, Error: "episodes must be between 1 and 63"})
			return
		}
		if errors.Is(err, model.ErrNotFound) {
			jsonEncoder(w, http.StatusNotFound, response{Data: nil, Error: "narration not found"})
			return
		}
		log.Printf("failed to get narrations: episodes=%d lang=%s err=%v", episodes, lang, err)
		jsonEncoder(w, http.StatusInternalServerError, response{Data: nil, Error: "internal server error"})
		return
	}
	writeNarrations(w, narrations, lang)
}

func (h *NarrationHandler) OneNarrationHandler(w http.ResponseWriter, r *http.Request) {
	lang, ok := resolveLang(r)
	if !ok {
		jsonEncoder(w, http.StatusBadRequest, response{Data: nil, Error: "the language is not supported"})
		return
	}

	// パスパラメーターは文字列で届くため、数値でなければリクエスト不正として扱う。
	episode, err := strconv.Atoi(r.PathValue("episode"))
	if err != nil {
		jsonEncoder(w, http.StatusBadRequest, response{Data: nil, Error: "episode must be an integer"})
		return
	}

	narration, err := h.service.GetNarrationByEpisode(episode, lang)
	if err != nil {
		// 未翻訳もエピソード不在も、クライアントからは区別せず404で返す。
		if errors.Is(err, model.ErrNotFound) {
			jsonEncoder(w, http.StatusNotFound, response{Data: nil, Error: "narration not found"})
			return
		}
		log.Printf("failed to get narration: episode=%d lang=%s err=%v", episode, lang, err)
		jsonEncoder(w, http.StatusInternalServerError, response{Data: nil, Error: "internal server error"})
		return
	}

	writeNarration(w, narration, lang)
}

func jsonEncoder(w http.ResponseWriter, statusCode int, res any) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(res)
}
