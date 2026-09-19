package handler

import (
	"encoding/json"
	"fullmetal-api/internal/model"
	"net/http"
)

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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response{Data: nil, Error: "the language is not supported"})
		return
	}

	narration := h.service.GetRandomNarration(lang)
	if narration == nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response{})
		return
	}

	res := narrationResponse{
		Episode:    narration.Episode,
		Title:      *narration.Title[lang], // サービスレイヤーで非nilのものに絞り込み済み
		Narrations: narration.Narrations[lang],
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response{Data: &res})
}
