package main

import (
	"context"
	"gin-api-1/internal/env"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v3"
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
	resend *resend.Client
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

	client := resend.NewClient(env.GetEnvString("RESEND_API_KEY", "re_xx"))

	app := &application{
		config: cfg,
		db:     conn,
		resend: client,
	}

	err := app.serve(app.routes())

	if err != nil {
		log.Fatal(err)
	}
}
