package main

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	url := strings.TrimSpace(os.Getenv("REDIS_URL"))
	if url == "" {
		fmt.Println("REDIS_URL ausente")
		os.Exit(2)
	}
	opt, err := redis.ParseURL(url)
	if err != nil {
		fmt.Println("parse:", err)
		os.Exit(1)
	}
	c := redis.NewClient(opt)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := c.Ping(ctx).Err(); err != nil {
		fmt.Println("PING FAIL:", err)
		os.Exit(1)
	}
	fmt.Println("PING OK")
	for _, k := range []string{"confvision:eventos", "confvision:eventos:dlq"} {
		n, err := c.LLen(ctx, k).Result()
		if err != nil {
			fmt.Printf("LLEN %s FAIL: %v\n", k, err)
			continue
		}
		fmt.Printf("LLEN %s = %d\n", k, n)
	}
	keys, _ := c.Keys(ctx, "confvision:*").Result()
	fmt.Printf("KEYS confvision:* count=%d\n", len(keys))
	if len(keys) > 0 && len(keys) <= 20 {
		for _, k := range keys {
			fmt.Println(" ", k)
		}
	}
}
