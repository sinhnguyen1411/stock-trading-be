package users

import (
	"context"

	"github.com/sinhnguyen1411/stock-trading-be/api/grpc/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserService) Delete(ctx context.Context, req *user.DeleteRequest) (*user.DeleteResponse, error) {
	if err := s.deleteUseCase.DeleteAccount(ctx, req.Username); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete account: %v", err)
	}
	return &user.DeleteResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
	}, nil
}
