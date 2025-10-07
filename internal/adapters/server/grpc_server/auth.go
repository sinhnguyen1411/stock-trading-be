package grpc_server

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	userpb "github.com/sinhnguyen1411/stock-trading-be/api/grpc/user"
	"github.com/sinhnguyen1411/stock-trading-be/internal/security"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthUnaryServerInterceptor validates Authorization: Bearer <token> for protected methods.
func AuthUnaryServerInterceptor(tokenValidator security.AccessTokenManager) grpc.UnaryServerInterceptor {
	// Public methods that do not require authentication
	public := map[string]struct{}{
		userpb.UserService_Login_FullMethodName:              {},
		userpb.UserService_Register_FullMethodName:           {},
		userpb.UserService_ResendVerification_FullMethodName: {},
		userpb.UserService_VerifyUser_FullMethodName:         {},
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		if _, ok := public[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}
		var authz string
		if vals := md.Get("authorization"); len(vals) > 0 {
			authz = vals[0]
		}
		if authz == "" {
			return nil, status.Error(codes.Unauthenticated, "missing authorization header")
		}

		parts := strings.SplitN(authz, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
		}
		token := strings.TrimSpace(parts[1])
		if token == "" {
			return nil, status.Error(codes.Unauthenticated, "empty token")
		}

		claims, err := tokenValidator.ValidateAccessToken(token)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("invalid token: %v", err))
		}

		ctx = context.WithValue(ctx, ctxKeyClaims{}, claims)
		slog.Debug("AUTH OK", "method", info.FullMethod, "uid", claims.UserID)

		return handler(ctx, req)
	}
}

type ctxKeyClaims struct{}

// UserIDFromContext returns the user ID parsed from Authorization token by the interceptor.
func UserIDFromContext(ctx context.Context) (int64, bool) {
	v := ctx.Value(ctxKeyClaims{})
	if v == nil {
		return 0, false
	}
	claims, ok := v.(*security.AccessTokenClaims)
	if !ok {
		return 0, false
	}
	return claims.UserID, true
}

// ClaimsFromContext exposes full token claims for handlers that need more details.
func ClaimsFromContext(ctx context.Context) (*security.AccessTokenClaims, bool) {
	v := ctx.Value(ctxKeyClaims{})
	if v == nil {
		return nil, false
	}
	claims, ok := v.(*security.AccessTokenClaims)
	return claims, ok
}
