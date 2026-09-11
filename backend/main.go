package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type Submission struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	Name        string    `json:"name" gorm:"size:100;not null"`
	Phone       string    `json:"phone" gorm:"size:20;not null"`
	SubmittedAt time.Time `json:"submittedAt"`
}

type createRequest struct {
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type errorResponse struct {
	Error string `json:"error"`
}

var cnMobile = regexp.MustCompile(`^1[3-9]\d{9}$`)

func main() {
	dbPath := getenv("DB_PATH", "data.db")
	addr := listenAddr()
	origins := parseCORSOrigins(os.Getenv("CORS_ORIGINS"))

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		log.Fatalf("create db dir: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	if err := db.AutoMigrate(&Submission{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("POST /api/submissions", func(w http.ResponseWriter, r *http.Request) {
		createSubmission(w, r, db)
	})
	mux.HandleFunc("GET /api/submissions", func(w http.ResponseWriter, r *http.Request) {
		listSubmissions(w, r, db)
	})

	handler := corsMiddleware(origins, mux)

	if len(origins) == 0 {
		log.Printf("API listening on %s (db=%s, cors=*)", addr, dbPath)
	} else {
		log.Printf("API listening on %s (db=%s, cors=%s)", addr, dbPath, strings.Join(origins, ","))
	}
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func createSubmission(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "请求格式无效")
		return
	}

	name := strings.TrimSpace(req.Name)
	phone := strings.TrimSpace(req.Phone)

	if name == "" {
		writeError(w, http.StatusBadRequest, "姓名不能为空")
		return
	}
	if !cnMobile.MatchString(phone) {
		writeError(w, http.StatusBadRequest, "请输入有效的中国大陆手机号")
		return
	}

	item := Submission{
		Name:        name,
		Phone:       phone,
		SubmittedAt: time.Now().UTC(),
	}
	if err := db.Create(&item).Error; err != nil {
		log.Printf("create submission: %v", err)
		writeError(w, http.StatusInternalServerError, "保存失败，请稍后重试")
		return
	}

	writeJSON(w, http.StatusCreated, item)
}

func listSubmissions(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	var items []Submission
	if err := db.Order("submitted_at DESC, id DESC").Find(&items).Error; err != nil {
		log.Printf("list submissions: %v", err)
		writeError(w, http.StatusInternalServerError, "读取失败，请稍后重试")
		return
	}
	if items == nil {
		items = []Submission{}
	}
	writeJSON(w, http.StatusOK, items)
}

func corsMiddleware(origins []string, next http.Handler) http.Handler {
	allowAll := len(origins) == 0
	allowed := make(map[string]struct{}, len(origins))
	for _, origin := range origins {
		allowed[origin] = struct{}{}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if allowAll {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin := r.Header.Get("Origin"); origin != "" {
			if _, ok := allowed[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// listenAddr prefers Render's PORT, then ADDR, then :8080.
func listenAddr() string {
	if port := strings.TrimSpace(os.Getenv("PORT")); port != "" {
		if strings.HasPrefix(port, ":") {
			return port
		}
		return ":" + port
	}
	if addr := strings.TrimSpace(os.Getenv("ADDR")); addr != "" {
		return addr
	}
	return ":8080"
}

func parseCORSOrigins(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "*" {
		return nil
	}
	var origins []string
	for _, part := range strings.Split(raw, ",") {
		origin := strings.TrimRight(strings.TrimSpace(part), "/")
		if origin != "" && origin != "*" {
			origins = append(origins, origin)
		}
	}
	return origins
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, errorResponse{Error: msg})
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
