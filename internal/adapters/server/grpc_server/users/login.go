package users

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/sinhnguyen1411/stock-trading-be/api/grpc/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserService) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	accessToken, accessExpires, refreshToken, refreshExpires, info, err := s.loginUseCase.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		errLower := strings.ToLower(err.Error())
		if strings.Contains(errLower, "invalid credentials") {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		return nil, status.Errorf(codes.Internal, "login failed: %v", err)
	}

	payload := map[string]interface{}{
		"name":                          info.Name,
		"email":                         info.Email,
		"cmnd":                          info.DocumentID,
		"birthday":                      info.Birthday.Unix(),
		"gender":                        info.Gender,
		"permanent_address":             info.PermanentAddress,
		"phone_number":                  info.PhoneNumber,
		"token_expires_at_unix":         accessExpires.Unix(),
		"refresh_token_expires_at_unix": refreshExpires.Unix(),
	}
	body, _ := json.Marshal(payload)

	return &user.LoginResponse{
		Code:    uint32(codes.OK),
		Message: string(body),
		Data: &user.LoginResponse_Data{
			Token:                     accessToken,
			TokenExpiresAtUnix:        accessExpires.Unix(),
			RefreshToken:              refreshToken,
			RefreshTokenExpiresAtUnix: refreshExpires.Unix(),
		},
	}, nil
}
