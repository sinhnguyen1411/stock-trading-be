package users

import (
	"context"
	"strings"

	"github.com/sinhnguyen1411/stock-trading-be/api/grpc/user"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *UserService) RefreshToken(ctx context.Context, req *user.RefreshTokenRequest) (*user.RefreshTokenResponse, error) {
	result, err := s.refreshUseCase.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		errLower := strings.ToLower(err.Error())
		switch {
		case strings.Contains(errLower, "validate refresh token"), strings.Contains(errLower, "token expired"), strings.Contains(errLower, "token revoked"), strings.Contains(errLower, "token invalid"):
			return nil, status.Error(codes.Unauthenticated, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "refresh token: %v", err)
		}
	}

	return &user.RefreshTokenResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
		Data: &user.LoginResponse_Data{
			Token:                     result.AccessToken,
			TokenExpiresAtUnix:        result.AccessTokenExpires.Unix(),
			RefreshToken:              result.RefreshToken,
			RefreshTokenExpiresAtUnix: result.RefreshTokenExpires.Unix(),
		},
	}, nil
}
