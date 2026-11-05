package main

import (
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"clotesShop/auth"
)

func main() {
	a, _ := auth.HashPassword("popcorn2569")
	b, _ := auth.HashPassword("popcorn2569")
	fmt.Println(a)
	fmt.Println(b)

	fmt.Println("popcorn2569 ตรงกับ a:", auth.CheckPassword(a, "popcorn2569"))
	fmt.Println("popcorn2569 ตรงกับ b:", auth.CheckPassword(b, "popcorn2569"))
	fmt.Println("popcorn2568 ตรงกับ a:", auth.CheckPassword(a, "popcorn2568"))

	for _, cost := range []int{8, 10, 12} {
		start := time.Now()
		bcrypt.GenerateFromPassword([]byte("popcorn2569"), cost)
		fmt.Printf("cost %d ใช้เวลา %v\n", cost, time.Since(start).Round(time.Millisecond))
	}

	_, err := auth.HashPassword(strings.Repeat("a", 73))
	fmt.Println("73 ไบต์:", err)
}
