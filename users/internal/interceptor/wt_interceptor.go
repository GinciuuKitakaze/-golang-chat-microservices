package interceptor

import (
	"context"
	"runtime/debug"
	"time"
	"uuid"

	core_logger "github.com/GinciuuKitakaze/users/internal/core/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ServerInterceptor struct {
	logger *core_logger.Logger
}

func NewServerInterceptor(logger *core_logger.Logger) *ServerInterceptor {
	return &ServerInterceptor{
		logger: logger,
	}
}

func (s *ServerInterceptor) Recovery() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				s.logger.Error("panic recovered",
					zap.String("method", info.FullMethod),
					zap.Any("panic", r),
					zap.ByteString("stack", debug.Stack()))
				err = status.Error(codes.Internal, "internal server error")
			}
		}()

		return handler(ctx, req)
	}
}

func (s *ServerInterceptor) RequestID() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		var requestID string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get("x-request-id"); len(values) > 0 {
				requestID = values[0]
			}
		}

		if requestID == "" {
			requestID = uuid.NewV7().String()
		}

		ctx = context.WithValue(ctx, requestIDKey, requestID)

		ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", requestID)
		return handler(ctx, req)
	}
}

func (s *ServerInterceptor) Logging() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		fields := []zap.Field{
			zap.String("method", info.FullMethod),
			zap.Duration("latency", time.Since(start)),
			zap.String("code", status.Code(err).String()),
			zap.String("request_id", RequestIDFromContext(ctx)),
		}

		if err != nil {
			s.logger.Warn("gRPC call failed", append(fields, zap.Error(err))...)
		} else {
			s.logger.Debug("gRPC call", fields...)
		}

		return resp, err
	}
}
