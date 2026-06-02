package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/mar1mo-41414/ore2ca/internal/ca"
	"github.com/mar1mo-41414/ore2ca/internal/config"
	"github.com/mar1mo-41414/ore2ca/internal/store"
)

type Server struct {
	store *store.Store
	cfg   *config.Config
	addr  string
}

func NewServer(addr string) (*Server, error) {
	s, err := store.New()
	if err != nil {
		return nil, err
	}
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}
	return &Server{store: s, cfg: cfg, addr: addr}, nil
}

func (srv *Server) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", srv.handleIndex)
	mux.HandleFunc("GET /api/ca", srv.handleCAInfo)
	mux.HandleFunc("GET /api/certs", srv.handleListCerts)
	mux.HandleFunc("POST /api/issue", srv.handleIssue)
	mux.HandleFunc("POST /api/revoke/{id}", srv.handleRevoke)
	mux.HandleFunc("DELETE /api/certs/{id}", srv.handleDelete)
	mux.HandleFunc("GET /download/{domain}/{file}", srv.handleDownload)

	fmt.Printf("ore2ca Web UI 起動中: http://%s\n", srv.addr)
	fmt.Println("停止するには Ctrl+C")
	return http.ListenAndServe(srv.addr, mux)
}

// --- API handlers ---

type caInfoResponse struct {
	CommonName   string `json:"common_name"`
	Organization string `json:"organization"`
	Country      string `json:"country"`
	CertPath     string `json:"cert_path"`
	Exists       bool   `json:"exists"`
}

func (srv *Server) handleCAInfo(w http.ResponseWriter, r *http.Request) {
	resp := caInfoResponse{
		CommonName:   srv.cfg.CA.CommonName,
		Organization: srv.cfg.CA.Organization,
		Country:      srv.cfg.CA.Country,
		CertPath:     srv.store.CACertPath(),
		Exists:       srv.store.CAExists(),
	}
	jsonOK(w, resp)
}

func (srv *Server) handleListCerts(w http.ResponseWriter, r *http.Request) {
	metas, err := srv.store.ListCerts()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if metas == nil {
		metas = []*store.CertMeta{}
	}
	jsonOK(w, metas)
}

type issueRequest struct {
	Domain    string   `json:"domain"`
	ExtraSANs []string `json:"extra_sans"`
	Days      int      `json:"days"`
}

func (srv *Server) handleIssue(w http.ResponseWriter, r *http.Request) {
	var req issueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "リクエスト解析エラー: "+err.Error(), http.StatusBadRequest)
		return
	}
	if req.Domain == "" {
		jsonError(w, "domain は必須です", http.StatusBadRequest)
		return
	}
	cfg := *srv.cfg
	if req.Days > 0 {
		cfg.Certs.ValidDays = req.Days
	}
	meta, err := ca.Issue(req.Domain, &cfg, srv.store, req.ExtraSANs...)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, meta)
}

func (srv *Server) handleRevoke(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	meta, err := srv.store.FindByID(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	if meta.Revoked {
		jsonError(w, "すでに失効済みです", http.StatusBadRequest)
		return
	}
	meta.Revoked = true
	if err := srv.store.SaveMeta(meta); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, meta)
}

func (srv *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	meta, err := srv.store.FindByID(id)
	if err != nil {
		jsonError(w, err.Error(), http.StatusNotFound)
		return
	}
	if err := srv.store.DeleteCert(meta.Domain); err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, map[string]string{"status": "deleted"})
}

func (srv *Server) handleDownload(w http.ResponseWriter, r *http.Request) {
	domain := r.PathValue("domain")
	file := r.PathValue("file")

	allowed := map[string]bool{
		"cert.crt":      true,
		"cert.key":      true,
		"chain.crt":     true,
		"fullchain.crt": true,
	}
	if !allowed[file] {
		http.Error(w, "不正なファイル名", http.StatusBadRequest)
		return
	}

	// sanitize: prevent path traversal
	if strings.Contains(domain, "..") || strings.Contains(domain, "/") {
		http.Error(w, "不正なドメイン名", http.StatusBadRequest)
		return
	}

	path := filepath.Join(srv.store.CertDir(domain), file)
	data, err := os.ReadFile(path)
	if err != nil {
		http.Error(w, "ファイルが見つかりません", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, file))
	w.Header().Set("Content-Type", "application/x-pem-file")
	w.Write(data)
}

// --- index (SPA shell) ---

func (srv *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(indexHTML))
}

// --- helpers ---

func jsonOK(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
