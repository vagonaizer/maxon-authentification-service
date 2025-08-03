package interceptors

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/vagonaizer/authenitfication-service/pkg/logger"
)

type LoggingInterceptor struct {
	logger *logger.Logger
}

func NewLoggingInterceptor(logger *logger.Logger) *LoggingInterceptor {
	return &LoggingInterceptor{
		logger: logger,
	}
}

func (i *LoggingInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()

		resp, err := handler(ctx, req)

		duration := time.Since(start)
		statusCode := status.Code(err)

		fields := logger.Fields{
			"method":   info.FullMethod,
			"duration": duration.String(),
			"status":   statusCode.String(),
		}

		if userID := ctx.Value("user_id"); userID != nil {
			fields["user_id"] = userID
		}

		if err != nil {
			fields["error"] = err.Error()
			i.logger.WithFields(fields).Error("grpc request completed with error")
		} else {
			i.logger.WithFields(fields).Info("grpc request completed")
		}

		return resp, err
	}
}

func (i *LoggingInterceptor) Stream() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()

		err := handler(srv, ss)

		duration := time.Since(start)
		statusCode := status.Code(err)

		fields := logger.Fields{
			"method":   info.FullMethod,
			"duration": duration.String(),
			"status":   statusCode.String(),
		}

		if userID := ss.Context().Value("user_id"); userID != nil {
			fields["user_id"] = userID
		}

		if err != nil {
			fields["error"] = err.Error()
			i.logger.WithFields(fields).Error("grpc stream completed with error")
		} else {
			i.logger.WithFields(fields).Info("grpc stream completed")
		}

		return err
	}
}
