package router

import (
	"net/http"

	"fullmetal-api/internal/handler"
)

func New(h *handler.NarrationHandler) *http.ServeMux {
	mux := http.NewServeMux()
	registerV1(mux, h)
	registerLegacyRedirects(mux)
	return mux
}

func registerV1(mux *http.ServeMux, h *handler.NarrationHandler) {
	mux.HandleFunc("GET /v1/narrations/random", h.NarrationHandler)
	mux.HandleFunc("GET /v1/narrations/{episode}", h.OneNarrationHandler)
}

// バージョン導入前に公開していた /narrations/... を壊さないための後方互換。
// 新規のパスは追加せず、既存URLの利用者が /v1 へ移行し次第削除する。
func registerLegacyRedirects(mux *http.ServeMux) {
	mux.HandleFunc("GET /narrations/{path...}", redirectToV1)
}

func redirectToV1(w http.ResponseWriter, r *http.Request) {
	target := "/v1" + r.URL.Path
	if r.URL.RawQuery != "" {
		target += "?" + r.URL.RawQuery
	}
	http.Redirect(w, r, target, http.StatusPermanentRedirect)
}
