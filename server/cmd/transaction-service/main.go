package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"go-core-banking-system/internal/transaction/handler"
	"go-core-banking-system/internal/transaction/repository"
	"go-core-banking-system/internal/transaction/service"
	"go-core-banking-system/pkg/proto/account"
	"go-core-banking-system/pkg/proto/transaction"
)

func main() {
	dbHost := getEnv("DB_HOST", "localhost")
	dbPort := getEnv("DB_PORT", "5432")
	dbUser := getEnv("DB_USER", "bank")
	dbPass := getEnv("DB_PASSWORD", "bank_secret")
	dbName := getEnv("DB_NAME", "transaction_db")
	grpcPort := getEnv("GRPC_PORT", "9002")
	accountAddr := getEnv("ACCOUNT_SERVICE_ADDR", "localhost:9001")

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPass, dbHost, dbPort, dbName)

	ctx := context.Background()

	runMigrations(dsn, "file://migrations/transaction")

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

	conn, err := grpc.NewClient(accountAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to account service: %v", err)
	}
	defer conn.Close()

	accountClient := account.NewAccountServiceClient(conn)
	accountCli := &grpcAccountClient{client: accountClient}

	repo := repository.NewPostgresRepo(pool)
	svc := service.NewTransactionService(repo, accountCli)
	grpcHandler := handler.NewTransactionGRPCHandler(svc)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	transaction.RegisterTransactionServiceServer(grpcServer, grpcHandler)

	log.Printf("Transaction Service listening on :%s", grpcPort)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type grpcAccountClient struct {
	client account.AccountServiceClient
}

func (c *grpcAccountClient) Reserve(ctx context.Context, id string, amount int64) error {
	_, err := c.client.Reserve(ctx, &account.ReserveRequest{Id: id, Amount: amount})
	return err
}

func (c *grpcAccountClient) Credit(ctx context.Context, id string, amount int64) error {
	_, err := c.client.Credit(ctx, &account.CreditRequest{Id: id, Amount: amount})
	return err
}

func (c *grpcAccountClient) Debit(ctx context.Context, id string, amount int64) error {
	_, err := c.client.Debit(ctx, &account.DebitRequest{Id: id, Amount: amount})
	return err
}

func (c *grpcAccountClient) CommitReservation(ctx context.Context, id string, amount int64) error {
	_, err := c.client.CommitReservation(ctx, &account.CommitReservationRequest{Id: id, Amount: amount})
	return err
}

func (c *grpcAccountClient) CancelReservation(ctx context.Context, id string, amount int64) error {
	_, err := c.client.CancelReservation(ctx, &account.CancelReservationRequest{Id: id, Amount: amount})
	return err
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func runMigrations(dsn, migrationsPath string) {
	m, err := migrate.New(migrationsPath, dsn)
	if err != nil {
		log.Fatalf("failed to create migrate instance: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run migrations: %v", err)
	}

	log.Println("Migrations applied successfully")
}
