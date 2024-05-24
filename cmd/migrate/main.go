package main

import (
	"context"

	"github.com/txix-open/isp-kit/db"
)

func main() {
	db.Open(context.Background(), "")
}
