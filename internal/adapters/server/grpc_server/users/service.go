package users

import (
	"context"
	"github.com/sinhnguyen1411/stock-trading-be/api/grpc/user"
	user2 "github.com/sinhnguyen1411/stock-trading-be/internal/usecases/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	user.UnimplementedUserServiceServer
	userUseCase user2.UserRegisterUseCase
}

func NewUserService(
	registerUseCase user2.UserRegisterUseCase,
) *UserService {
	return &UserService{
		userUseCase: registerUseCase,
	}
}

func (s *UserService) RegisterService(server grpc.ServiceRegistrar) {
	user.RegisterUserServiceServer(server, s)
}

func (s *UserService) Register(ctx context.Context, req *user.RegisterRequest) (*user.RegisterResponse, error) {
	err := s.userUseCase.RegisterAccount(ctx, user2.RequestRegister{
		Username:         req.Username,
		Password:         req.Password,
		Email:            req.Email,
		Name:             req.Name,
		Cmnd:             req.Cmnd,
		Birthday:         req.Birthday,
		Gender:           req.Gender,
		PermanentAddress: req.PermanentAddress,
		PhoneNumber:      req.PhoneNumber,
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register account: %v", err)
	}
	return &user.RegisterResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
	}, nil
}
