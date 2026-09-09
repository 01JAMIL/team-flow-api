package main

import (
	"context"
	"fmt"
	"gin-api-1/internal/adapters/postgresql/migrations"
	"gin-api-1/internal/cache"
	"gin-api-1/internal/env"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v3"
	"github.com/stripe/stripe-go/v86"
)

type databaseConfig struct {
	dsn string
}

type config struct {
	port      int
	db        databaseConfig
	redisHost string
	redisPort int
}

type application struct {
	config config
	db     *pgxpool.Pool
	resend *resend.Client
	stripe *stripe.Client
	cache  *cache.RedisCache
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
		redisHost: env.GetEnvString("REDIS_HOST", "localhost"),
		redisPort: env.GetEnvInt("REDIS_PORT", 6379),
	}

	ctx := context.Background()

	dbpool, poolErr := pgxpool.New(ctx, cfg.db.dsn)
	if poolErr != nil {
		panic(poolErr)
	}
	defer dbpool.Close()

	if err := dbpool.Ping(ctx); err != nil {
		panic(err)
	}

	log.Println("Database connected successfully")

	if err := migrations.Run(ctx, cfg.db.dsn); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Println("Database migrations applied successfully")

	client := resend.NewClient(env.GetEnvString("RESEND_API_KEY", "re_xx"))
	sc := stripe.NewClient(env.GetEnvString("STRIPE_SECRET_KEY", "stripe_xx"))

	redisAddr := fmt.Sprintf("%s:%d", cfg.redisHost, cfg.redisPort)
	redisCache := cache.NewRedisCache(redisAddr, 0, 5*time.Minute)

	log.Println("Redis cache initialized successfully")

	app := &application{
		config: cfg,
		db:     dbpool,
		resend: client,
		stripe: sc,
		cache:  redisCache,
	}

	err := app.serve(app.routes())

	if err != nil {
		log.Fatal(err)
	}
}
