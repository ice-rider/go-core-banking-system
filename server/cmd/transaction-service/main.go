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
	"google.golang.org/grpc/credentials/insecure"

	"go-core-banking-system/internal/transaction/handler"
	"go-core-banking-system/internal/transaction/repository"
	"go-core-banking-system/internal/transaction/service"
	"go-core-banking-system/pkg/app"
	"go-core-banking-system/pkg/observability"
	"go-core-banking-system/pkg/resilience"
	"go-core-banking-system/pkg/proto/account"
	"go-core-banking-system/pkg/proto/transaction"
)

func main() {
	dbHost := app.GetEnv("DB_HOST", "localhost")
	dbPort := app.GetEnv("DB_PORT", "5432")
	dbUser := app.GetEnv("DB_USER", "bank")
	dbPass := app.GetEnv("DB_PASSWORD", "bank_secret")
	dbName := app.GetEnv("DB_NAME", "transaction_db")
	grpcPort := app.GetEnv("GRPC_PORT", "9002")
	accountAddr := app.GetEnv("ACCOUNT_SERVICE_ADDR", "localhost:9001")
	otlpEndpoint := app.GetEnv("OTEL_EXPORTER_OTLP_ENDPOINT", "")

	shutdown, err := observability.Init("transaction-service", otlpEndpoint)
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

	app.RunMigrations(dsn, "file://migrations/transaction")

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

	accountCB := resilience.New(resilience.DefaultConfig("account-service"))

	conn, err := grpc.NewClient(accountAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithChainUnaryInterceptor(
			resilience.UnaryClientInterceptor(accountCB, 5*time.Second),
		),
	)
	if err != nil {
		log.Fatalf("failed to connect to account service: %v", err)
	}
	defer func() { _ = conn.Close() }()

	accountClient := account.NewAccountServiceClient(conn)
	accountCli := &grpcAccountClient{client: accountClient}

	repo := repository.NewPostgresRepo(pool)
	svc := service.NewTransactionService(repo, accountCli, nil)
	grpcHandler := handler.NewTransactionGRPCHandler(svc)

	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	unaryInterceptor, streamInterceptor := observability.GRPCServerInterceptors()
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(unaryInterceptor),
		grpc.ChainStreamInterceptor(streamInterceptor),
	)
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
