package users

import (
	"context"

	"github.com/sinhnguyen1411/stock-trading-be/api/grpc/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserService) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	token, err := s.loginUseCase.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}
	return &user.LoginResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),

		Data: &user.LoginResponse_Data{Token: token},
	}, nil
}
