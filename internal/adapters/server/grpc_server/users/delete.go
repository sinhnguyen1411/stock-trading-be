package users

import (
    "context"
    "errors"
    "strings"

    "github.com/sinhnguyen1411/stock-trading-be/api/grpc/user"
    grpcserver "github.com/sinhnguyen1411/stock-trading-be/internal/adapters/server/grpc_server"
    useruc "github.com/sinhnguyen1411/stock-trading-be/internal/usecases/user"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/metadata"
    "google.golang.org/grpc/status"
)

func (s *UserService) Delete(ctx context.Context, req *user.DeleteRequest) (*user.DeleteResponse, error) {
    username := req.GetUsername()

    // Fallback: if username is empty, try to parse from raw path metadata
    if username == "" {
        if md, ok := metadata.FromIncomingContext(ctx); ok {
            if paths := md.Get("path"); len(paths) > 0 {
                p := paths[0]
                // strip query string if any
                if i := strings.IndexByte(p, '?'); i >= 0 {
                    p = p[:i]
                }
                p = strings.TrimSuffix(p, "/")
                if idx := strings.LastIndexByte(p, '/'); idx >= 0 && idx < len(p)-1 {
                    username = p[idx+1:]
                }
            }
        }
    }

    if username == "" {
        return nil, status.Errorf(codes.InvalidArgument, "username is empty")
    }

    // Ensure the caller owns this account: compare uid from token with username's uid
    if uid, ok := grpcserver.UserIDFromContext(ctx); ok {
        if err := s.deleteUseCase.DeleteAccountOwned(ctx, uid, username); err != nil {
            switch {
            case errors.Is(err, useruc.ErrEmptyUsername):
                return nil, status.Error(codes.InvalidArgument, err.Error())
            case errors.Is(err, useruc.ErrPermissionDenied):
                return nil, status.Error(codes.PermissionDenied, err.Error())
            default:
                return nil, status.Errorf(codes.Internal, "failed to delete account: %v", err)
            }
        }
    } else {
        return nil, status.Error(codes.Unauthenticated, "unauthenticated")
    }
    return &user.DeleteResponse{
        Code:    uint32(codes.OK),
        Message: codes.OK.String(),
    }, nil
}
