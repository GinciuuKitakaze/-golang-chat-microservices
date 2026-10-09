package grpctransport

import (
	"errors"

	core_errors "github.com/GinciuuKitakaze/auth/internal/core/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func toGRPCError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, core_errors.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, core_errors.ErrValidation):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, core_errors.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, core_errors.ErrUnauthenticated):
		return status.Error(codes.Unauthenticated, "unauthenticated")
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
