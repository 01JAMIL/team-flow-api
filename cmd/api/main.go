package main

import (
	"context"
	"gin-api-1/internal/adapters/postgresql/migrations"
	"gin-api-1/internal/env"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v3"
	"github.com/stripe/stripe-go/v86"
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
	stripe *stripe.Client
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

	if err := migrations.Run(ctx, cfg.db.dsn); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("Database migrations applied successfully")

	client := resend.NewClient(env.GetEnvString("RESEND_API_KEY", "re_xx"))
	sc := stripe.NewClient(env.GetEnvString("STRIPE_SECRET_KEY", "stripe_xx"))

	app := &application{
		config: cfg,
		db:     conn,
		resend: client,
		stripe: sc,
	}

	err := app.serve(app.routes())

	if err != nil {
		log.Fatal(err)
	}
}
