package main

import (
	"fmt"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	h, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
	fmt.Print(string(h))
}
