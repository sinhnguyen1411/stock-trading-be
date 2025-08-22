package users

import (
	"context"

	"github.com/sinhnguyen1411/stock-trading-be/api/grpc/user"
	"google.golang.org/grpc/codes"
)

func (s *UserService) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	//do something
	return &user.LoginResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
		Data:    &user.LoginResponse_Data{Token: "alice123"},
	}, nil
}
