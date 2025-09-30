package users

import (
	"context"
	"strings"

	"github.com/sinhnguyen1411/stock-trading-be/api/grpc/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserService) Logout(ctx context.Context, req *user.LogoutRequest) (*user.LogoutResponse, error) {
	if err := s.logoutUseCase.Logout(ctx, req.GetRefreshToken()); err != nil {
		errLower := strings.ToLower(err.Error())
		switch {
		case strings.Contains(errLower, "token expired"), strings.Contains(errLower, "token invalid"):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "logout: %v", err)
		}
	}
	return &user.LogoutResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
	}, nil
}
