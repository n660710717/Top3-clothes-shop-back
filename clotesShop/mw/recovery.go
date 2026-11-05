package mw

import (
	"encoding/json"
	"log"
	"net/http"
)

// Recovery กัน panic ไม่ให้เซิร์ฟเวอร์ล้ม แล้วตอบข้อผิดพลาดภายในแบบเดียวกับ handler
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("internal panic: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]string{
						"code":    "INTERNAL",
						"message": "ระบบขัดข้อง ลองใหม่อีกครั้ง",
					},
				})
			}
		}()
		next.ServeHTTP(w, r)
	})
}
