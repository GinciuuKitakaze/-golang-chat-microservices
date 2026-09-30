package grpctransport

import (
	"errors"

	core_errors "github.com/GinciuuKitakaze/users/internal/core/errors"
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
	case errors.Is(err, core_errors.ErrOptimisticLock):
		return status.Error(codes.Aborted, err.Error())
	case errors.Is(err, core_errors.ErrValidation):
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
