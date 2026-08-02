package main

import (
	"context"
	"gin-api-1/internal/env"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

type databaseConfig struct {
	dsn string
}

type config struct {
	port int
	db   databaseConfig
}

type application struct {
	config config
	db     *pgx.Conn
}

func main() {

	// load .env
	envErr := godotenv.Load()
	if envErr != nil {
		log.Println("No .env file found")
	}

	cfg := config{
		port: env.GetEnvInt("PORT", 8080),
		db: databaseConfig{
			dsn: env.GetFormattedDsn(),
		},
	}

	ctx := context.Background()
	conn, pgErr := pgx.Connect(ctx, cfg.db.dsn)
	if pgErr != nil {
		panic(pgErr)
	}
	defer conn.Close(ctx)

	log.Println("Database connected successfully")

	app := &application{
		config: cfg,
		db:     conn,
	}

	err := app.serve(app.routes())

	if err != nil {
		log.Fatal(err)
	}
}
