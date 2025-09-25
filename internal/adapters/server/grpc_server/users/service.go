package users

import (
	"context"
	"strings"

	"github.com/sinhnguyen1411/stock-trading-be/api/grpc/user"
	user2 "github.com/sinhnguyen1411/stock-trading-be/internal/usecases/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	user.UnimplementedUserServiceServer
	userUseCase   user2.UserRegisterUseCase
	loginUseCase  user2.UserLoginUseCase
	deleteUseCase user2.UserDeleteUseCase
	getUseCase    user2.UserGetUseCase
}

func NewUserService(
	registerUseCase user2.UserRegisterUseCase,
	loginUseCase user2.UserLoginUseCase,
	deleteUseCase user2.UserDeleteUseCase,
	getUseCase user2.UserGetUseCase,
) *UserService {
	return &UserService{
		userUseCase:   registerUseCase,
		loginUseCase:  loginUseCase,
		deleteUseCase: deleteUseCase,
		getUseCase:    getUseCase,
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

func (s *UserService) Get(ctx context.Context, req *user.GetUserRequest) (*user.GetUserResponse, error) {
	entity, err := s.getUseCase.Get(ctx, req.GetUsername())
	if err != nil {
		code := codes.Internal
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			code = codes.NotFound
		}
		return nil, status.Errorf(code, "get user: %v", err)
	}

	var (
		birthday int64
		created  int64
		updated  int64
	)
	if !entity.Birthday.IsZero() {
		birthday = entity.Birthday.Unix()
	}
	if !entity.CreatedAt.IsZero() {
		created = entity.CreatedAt.Unix()
	}
	if !entity.UpdatedAt.IsZero() {
		updated = entity.UpdatedAt.Unix()
	}

	return &user.GetUserResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
		Data: &user.UserProfile{
			Id:               entity.Id,
			Username:         req.GetUsername(),
			Name:             entity.Name,
			Email:            entity.Email,
			Cmnd:             entity.DocumentID,
			Birthday:         birthday,
			Gender:           entity.Gender,
			PermanentAddress: entity.PermanentAddress,
			PhoneNumber:      entity.PhoneNumber,
			CreatedAt:        created,
			UpdatedAt:        updated,
		},
	}, nil
}
