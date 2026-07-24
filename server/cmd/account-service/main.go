package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"

	"go-core-banking-system/internal/account/handler"
	"go-core-banking-system/internal/account/repository"
	"go-core-banking-system/internal/account/service"
	"go-core-banking-system/pkg/app"
	"go-core-banking-system/pkg/observability"
	"go-core-banking-system/pkg/proto/account"
)

func main() {
	dbHost := app.GetEnv("DB_HOST", "localhost")
	dbPort := app.GetEnv("DB_PORT", "5432")
	dbUser := app.GetEnv("DB_USER", "bank")
	dbPass := app.GetEnv("DB_PASSWORD", "bank_secret")
	dbName := app.GetEnv("DB_NAME", "account_db")
	grpcPort := app.GetEnv("GRPC_PORT", "9001")
	otlpEndpoint := app.GetEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	shutdown, err := observability.Init("account-service", otlpEndpoint)
	if err != nil {
		log.Printf("warning: observability init failed: %v", err)
	}
	if shutdown != nil {
		defer func() {
			_ = shutdown(context.Background())
		}()
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	ctx := context.Background()

	app.RunMigrations(dsn, "file://migrations/account")

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}

	cfg.MaxConns = 25
	cfg.MinConns = 5
	cfg.MaxConnLifetime = 5 * time.Minute
	cfg.MaxConnIdleTime = 3 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}

	repo := repository.NewPostgresRepo(pool)
	svc := service.NewAccountService(repo)
	grpcHandler := handler.NewAccountGRPCHandler(svc)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	unaryInterceptor, streamInterceptor := observability.GRPCServerInterceptors()
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptor),
		grpc.ChainStreamInterceptor(streamInterceptor),
	)
	account.RegisterAccountServiceServer(grpcServer, grpcHandler)

	log.Printf("Account Service listening on :%s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
