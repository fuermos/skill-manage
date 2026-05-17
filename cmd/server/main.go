package main

import (
	"flag"
	"log"

	"skill-manage/internal/auth"
	"skill-manage/internal/server"
	"skill-manage/internal/store"
)

func main() {
	addr := flag.String("addr", ":8080", "server listen address")
	dbPath := flag.String("db", "./server-data/sync.db", "database file path")
	token := flag.String("token", "change-me-in-production", "API auth token")
	flag.Parse()

	db, err := store.OpenDB(*dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	syncStore := store.NewSyncStore(db)
	skillStore := store.NewSkillStore(db)
	tokenAuth := auth.NewTokenAuth(*token)

	srv := server.NewServer(syncStore, skillStore, tokenAuth, *addr)

	log.Printf("Skill Sync Server starting on %s", *addr)
	log.Printf("Database: %s", *dbPath)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}