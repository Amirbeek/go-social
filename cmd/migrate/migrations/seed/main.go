package main

import (
	"log"

	"github.com/amirbeek/social/internal/db"
	"github.com/amirbeek/social/internal/env"
	"github.com/amirbeek/social/internal/store"
)

func main() {
	addr := env.GetString("DB_ADDR", "postgres://supervillager:adminpassword@localhost:5433/social?sslmode=disable")

	conn, err := db.New(addr, 30, 30, "15m")
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}
	defer conn.Close()

	storage := store.NewStorage(conn)
	db.Seed(storage, conn)

	log.Println("✅ Database seeded successfully!")
}
