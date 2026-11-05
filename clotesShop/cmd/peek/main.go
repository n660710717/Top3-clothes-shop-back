package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// peek เปิดอ่านตั๋วโดยไม่ใช้กุญแจลับเลยสักตัว
func main() {
	if len(os.Args) < 2 {
		fmt.Println("ใช้: peek <token>")
		return
	}
	parts := strings.Split(os.Args[1], ".")
	fmt.Println("ตั๋วมี", len(parts), "ท่อน")
	for i, name := range []string{"ส่วนหัว", "เนื้อความ"} {
		if i >= len(parts) {
			break
		}
		raw, err := base64.RawURLEncoding.DecodeString(parts[i])
		if err != nil {
			fmt.Println(name, "อ่านไม่ออก", err)
			continue
		}
		fmt.Printf("%s %s\n", name, raw)
	}
	if len(parts) > 2 {
		fmt.Println("ลายเซ็น", parts[2], "ยาว", len(parts[2]), "ตัว")
	}
}
