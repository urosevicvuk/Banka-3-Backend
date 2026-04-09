package main

import (
	"fmt"
	"log/slog"
	"net"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/RAF-SI-2025/Banka-3-Backend/gen/notification"
	internalNotification "github.com/RAF-SI-2025/Banka-3-Backend/internal/notification"
	"github.com/RAF-SI-2025/Banka-3-Backend/pkg/logger"
)

func main() {
	logger.Init("notification")

	port := os.Getenv("GRPC_PORT")
	if port == "" {
		port = "50051"
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		slog.Error("failed to listen", "port", port, "err", err)
		os.Exit(1)
	}
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(logger.UnaryServerInterceptor()),
		grpc.StreamInterceptor(logger.StreamServerInterceptor()),
	)
	smtpSender := &internalNotification.SMTPSender{}
	server := internalNotification.NewServer(smtpSender)

	notification.RegisterNotificationServiceServer(grpcServer, server)
	reflection.Register(grpcServer)
	slog.Info("notification service listening", "port", port)
	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("failed to serve", "err", err)
		os.Exit(1)
	}
}
