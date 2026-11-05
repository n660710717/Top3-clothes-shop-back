package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"clotesShop/auth"
	"clotesShop/handler"
	"clotesShop/mw"
	"clotesShop/store"
)

func main() {
	loadEnv(".env")

	secret := os.Getenv("JWT_SECRET")
	if len(secret) < 32 {
		log.Fatal("JWT_SECRET ต้องยาวอย่างน้อย 32 ตัวอักษร")
	}

	cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	cfg.MaxConns = 10
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	if err := pool.Ping(context.Background()); err != nil {
		log.Fatal("ต่อฐานข้อมูลไม่ติด ", err)
	}

	tickets := &auth.Issuer{Secret: []byte(secret), TTL: time.Hour}
	h := &handler.Handler{Store: &store.Store{Pool: pool}, Tickets: tickets}
	login := func(f http.HandlerFunc) http.HandlerFunc { return mw.RequireLogin(tickets, f) }
	seller := func(f http.HandlerFunc) http.HandlerFunc { return login(mw.RequireRole("seller", f)) }
	admin := func(f http.HandlerFunc) http.HandlerFunc { return login(mw.RequireRole("admin", f)) }

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/auth/register", h.Register)
	mux.HandleFunc("POST /api/auth/login", h.Login)
	mux.HandleFunc("GET /api/products", h.ListProducts)
	mux.HandleFunc("GET /api/products/{id}", h.GetProduct)
	mux.HandleFunc("GET /api/seller/products", seller(h.ListProducts))
	mux.HandleFunc("DELETE /api/seller/products/{id}", seller(h.DeleteProduct))
	mux.HandleFunc("GET /api/admin/members", admin(h.AdminMembers))
	mux.HandleFunc("PUT /api/admin/members/{memberId}/role", admin(h.SetMemberRole))

	originsEnv := os.Getenv("ALLOWED_ORIGINS")
	if strings.TrimSpace(originsEnv) == "" {
		originsEnv = "http://localhost:3000"
	}
	origins := strings.Split(originsEnv, ",")

	var app http.Handler = http.TimeoutHandler(mux, 10*time.Second, "ระบบใช้เวลานานเกินไป")
	app = mw.Logging(mw.Recovery(mw.CORS(origins, app)))

	srv := &http.Server{Addr: ":8080", Handler: app}
	go func() {
		log.Println("Clothes Shop API ฟังอยู่ที่ :8080 ต้อนรับ", origins)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("ได้รับสัญญาณปิดร้าน ไม่รับคำขอใหม่แล้ว")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Println("ปิดไม่ทัน", err)
	}
	log.Println("ปิดร้านเรียบร้อย")
}

// loadEnv อ่านไฟล์ .env เฉพาะคีย์ที่ยังไม่ได้ตั้งในสภาพแวดล้อม
func loadEnv(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		return
	}
	for _, line := range strings.Split(string(b), "\n") {
		line = strings.TrimSpace(strings.TrimRight(line, "\r"))
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if key != "" && os.Getenv(key) == "" {
			os.Setenv(key, val)
		}
	}
}
