package main

import (
	"database/sql"
	"fmt"
	"log/slog"
	"net"
	"os"

	"github.com/RAF-SI-2025/Banka-3-Backend/gen/bank"
	internalBank "github.com/RAF-SI-2025/Banka-3-Backend/internal/bank"
	"github.com/RAF-SI-2025/Banka-3-Backend/pkg/logger"
	_ "github.com/jackc/pgx/v5/stdlib"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func connect_to_db_gorm() *gorm.DB {
	dsn := os.Getenv("DATABASE_URL")
	gorm_db, gorm_err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if gorm_err != nil {
		slog.Error("gorm open failed", "err", gorm_err, "dsn", dsn)
		os.Exit(1)
	}
	return gorm_db
}

func connectToDB() *sql.DB {
	connStr := os.Getenv("DATABASE_URL")
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		slog.Error("sql open failed", "err", err)
		os.Exit(1)
	}
	return db
}

func main() {
	logger.Init("bank")

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		slog.Error("failed to listen", "port", port, "err", err)
		os.Exit(1)
	}

	db := connectToDB()
	gorm_db := connect_to_db_gorm()
	slog.Info("connected to database")
	defer func() { _ = db.Close() }()

	bankService, err := internalBank.NewServer(db, gorm_db)
	if err != nil {
		slog.Error("failed to start bank service", "err", err)
		os.Exit(1)
	}
	stopScheduler := bankService.StartScheduler()
	defer stopScheduler()

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(logger.UnaryServerInterceptor()),
		grpc.StreamInterceptor(logger.StreamServerInterceptor()),
	)
	bank.RegisterBankServiceServer(srv, bankService)
	reflection.Register(srv)

	slog.Info("bank service listening", "port", port)
	if err := srv.Serve(lis); err != nil {
		slog.Error("failed to serve", "err", err)
		os.Exit(1)
	}
}
