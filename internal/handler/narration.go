package handler

import (
	"encoding/json"
	"fullmetal-api/internal/model"
	"log"
	"net/http"
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

type NarrationService interface {
	GetRandomNarration(lang string) *model.Narration
}

type NarrationHandler struct {
	service NarrationService
}

func NewNarrationHandler(service NarrationService) *NarrationHandler {
	return &NarrationHandler{service: service}
}

func (h *NarrationHandler) NarrationHandler(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")

	// デフォルトでは日本語に設定することを担保
	if lang == "" {
		lang = "ja"
	}

	// Todo: 言語リストを作成し、パラメーターがリストに含まれるかの判定を行うようにする。
	if lang != "ja" && lang != "en" {
		jsonEncoder(w, http.StatusBadRequest, response{Data: nil, Error: "the language is not supported"})
		return
	}

	narration := h.service.GetRandomNarration(lang)
	if narration == nil {
		jsonEncoder(w, http.StatusOK, response{})
		return
	}

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

func jsonEncoder(w http.ResponseWriter, statusCode int, res response) {
	w.Header().Set("Content-Type", contentTypeJSON)
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(res)
}
