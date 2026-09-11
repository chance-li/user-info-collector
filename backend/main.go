package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
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
	addr := getenv("ADDR", ":8080")

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

	handler := corsMiddleware(mux)

	log.Printf("API listening on %s (db=%s)", addr, dbPath)
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

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
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
