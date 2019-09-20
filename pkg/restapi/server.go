package restapi

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path"
	"time"

	"github.com/cbsinteractive/drm-proxy/pkg/drmproxy"
	"github.com/google/uuid"
)

func NewServer(timeout time.Duration, port int, contentKeySvc drmproxy.KeyService) (*http.Server, error) {
	mux := http.NewServeMux()

	mux.Handle("/fairplay/", fairplayHandler{contentKeySvc: contentKeySvc})

	return &http.Server{
		Handler:      mux,
		Addr:         fmt.Sprintf(":%d", port),
		ReadTimeout:  timeout,
		WriteTimeout: timeout,
		IdleTimeout:  timeout,
	}, nil
}

type fairplayHandler struct {
	contentKeySvc drmproxy.KeyService
}

func (h fairplayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	reqID := uuid.New()
	log.Printf("started handling request url=%s, reqID=%s", r.URL, reqID)

	cfg := drmproxy.ContentCfg{
		ContentID: path.Base(path.Dir(r.URL.Path)),
		KeyID:     path.Base(r.URL.Path),
	}

	contentKeyResp, err := h.contentKeySvc.ContentKeyFrom(cfg)
	if err != nil {
		handleErr(r, w, err, reqID)
	}

	js, err := json.Marshal(contentKeyResp)
	if err != nil {
		handleErr(r, w, err, reqID)
	}

	log.Printf("finished handling request url=%s, reqID=%s", r.URL, reqID)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(js)
}

func handleErr(r *http.Request, w http.ResponseWriter, err error, reqID uuid.UUID) {
	log.Printf("error handling request url=%s err=%v reqID=%s", r.URL, err, reqID)
	w.WriteHeader(http.StatusInternalServerError)
	_, _ = w.Write([]byte(err.Error()))
}
