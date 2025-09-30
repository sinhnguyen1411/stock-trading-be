package users

import (
	"context"
	"errors"
	"strings"

	"github.com/sinhnguyen1411/stock-trading-be/api/grpc/user"
	user2 "github.com/sinhnguyen1411/stock-trading-be/internal/usecases/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	user.UnimplementedUserServiceServer
	userUseCase           user2.UserRegisterUseCase
	loginUseCase          user2.UserLoginUseCase
	deleteUseCase         user2.UserDeleteUseCase
	getUseCase            user2.UserGetUseCase
	updateUseCase         user2.UserUpdateUseCase
	changePasswordUseCase user2.UserChangePasswordUseCase
}

func NewUserService(
	registerUseCase user2.UserRegisterUseCase,
	loginUseCase user2.UserLoginUseCase,
	deleteUseCase user2.UserDeleteUseCase,
	getUseCase user2.UserGetUseCase,
	updateUseCase user2.UserUpdateUseCase,
	changePasswordUseCase user2.UserChangePasswordUseCase,
) *UserService {
	return &UserService{
		userUseCase:           registerUseCase,
		loginUseCase:          loginUseCase,
		deleteUseCase:         deleteUseCase,
		getUseCase:            getUseCase,
		updateUseCase:         updateUseCase,
		changePasswordUseCase: changePasswordUseCase,
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

func (s *UserService) Update(ctx context.Context, req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	err := s.updateUseCase.UpdateProfile(ctx, req.GetUsername(), user2.RequestUpdate{
		Email:            req.GetEmail(),
		Name:             req.GetName(),
		Cmnd:             req.GetCmnd(),
		Birthday:         req.GetBirthday(),
		Gender:           req.GetGender(),
		PermanentAddress: req.GetPermanentAddress(),
		PhoneNumber:      req.GetPhoneNumber(),
	})
	if err != nil {
		switch {
		case errors.Is(err, user2.ErrUpdateEmptyUsername), errors.Is(err, user2.ErrUpdateEmptyEmail), errors.Is(err, user2.ErrUpdateEmptyName):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case strings.Contains(strings.ToLower(err.Error()), "not found"):
			return nil, status.Errorf(codes.NotFound, "update user: %v", err)
		default:
			return nil, status.Errorf(codes.Internal, "update user: %v", err)
		}
	}

	return &user.UpdateUserResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
	}, nil
}

func (s *UserService) ChangePassword(ctx context.Context, req *user.ChangePasswordRequest) (*user.ChangePasswordResponse, error) {
	err := s.changePasswordUseCase.ChangePassword(ctx, req.GetUsername(), req.GetOldPassword(), req.GetNewPassword())
	if err != nil {
		switch {
		case errors.Is(err, user2.ErrChangePasswordEmptyUsername), errors.Is(err, user2.ErrChangePasswordEmptyNew):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, user2.ErrChangePasswordInvalidCurrent):
			return nil, status.Error(codes.PermissionDenied, err.Error())
		case strings.Contains(strings.ToLower(err.Error()), "not found"):
			return nil, status.Errorf(codes.NotFound, "change password: %v", err)
		default:
			return nil, status.Errorf(codes.Internal, "change password: %v", err)
		}
	}

	return &user.ChangePasswordResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
	}, nil
}
