package grpc_server

import (
    "context"
    "encoding/base64"
    "fmt"
    "strconv"
    "strings"

    userpb "github.com/sinhnguyen1411/stock-trading-be/api/grpc/user"
    "google.golang.org/grpc"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
)

// AuthUnaryServerInterceptor validates Authorization: Bearer <token> for protected methods.
func AuthUnaryServerInterceptor() grpc.UnaryServerInterceptor {
    // Public methods that do not require authentication
    public := map[string]struct{}{
        userpb.UserService_Login_FullMethodName:    {},
        userpb.UserService_Register_FullMethodName: {},
    }

    return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
        // Skip auth for public methods
        if _, ok := public[info.FullMethod]; ok {
            return handler(ctx, req)
        }

        // Extract Authorization header from metadata (lower-cased keys)
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

        // Expect format: Bearer <token>
        parts := strings.SplitN(authz, " ", 2)
        if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
            return nil, status.Error(codes.Unauthenticated, "invalid authorization header")
        }
        token := strings.TrimSpace(parts[1])
        if token == "" {
            return nil, status.Error(codes.Unauthenticated, "empty token")
        }

        // Basic validation: token must be base64 and contain userID:random
        if _, err := parseUserIDFromToken(token); err != nil {
            return nil, status.Error(codes.Unauthenticated, fmt.Sprintf("invalid token: %v", err))
        }

        // Optionally attach user id to context for handlers (not used currently)
        // ctx = context.WithValue(ctx, ctxKeyUserID{}, uid)

        return handler(ctx, req)
    }
}

// parseUserIDFromToken reproduces the token format from login use case: base64("<userID>:<random>")
func parseUserIDFromToken(token string) (int64, error) {
    decoded, err := base64.StdEncoding.DecodeString(token)
    if err != nil {
        return 0, fmt.Errorf("decode: %w", err)
    }
    s := string(decoded)
    idx := strings.IndexByte(s, ':')
    if idx <= 0 {
        return 0, fmt.Errorf("invalid format")
    }
    uidStr := s[:idx]
    uid, err := strconv.ParseInt(uidStr, 10, 64)
    if err != nil {
        return 0, fmt.Errorf("invalid user id")
    }
    return uid, nil
}
