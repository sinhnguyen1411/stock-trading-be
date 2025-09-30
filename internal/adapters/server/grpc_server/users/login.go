package users

import (
	"context"
	"encoding/json"

	"github.com/sinhnguyen1411/stock-trading-be/api/grpc/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserService) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	token, expiresAt, info, err := s.loginUseCase.Login(ctx, req.Username, req.Password)
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "login failed: %v", err)
	}

	payload := map[string]interface{}{
		"name":                  info.Name,
		"email":                 info.Email,
		"cmnd":                  info.DocumentID,
		"birthday":              info.Birthday.Unix(),
		"gender":                info.Gender,
		"permanent_address":     info.PermanentAddress,
		"phone_number":          info.PhoneNumber,
		"token_expires_at_unix": expiresAt.Unix(),
	}
	b, _ := json.Marshal(payload)

	return &user.LoginResponse{
		Code:    uint32(codes.OK),
		Message: string(b),
		Data:    &user.LoginResponse_Data{Token: token},
	}, nil
}
