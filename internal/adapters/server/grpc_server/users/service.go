package users

import (
	"context"
	"errors"
	"strings"

	"github.com/sinhnguyen1411/stock-trading-be/api/grpc/user"
	userentity "github.com/sinhnguyen1411/stock-trading-be/internal/entities/user"
	userusecase "github.com/sinhnguyen1411/stock-trading-be/internal/usecases/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserService struct {
	user.UnimplementedUserServiceServer
	registerUseCase           userusecase.UserRegisterUseCase
	resendVerificationUseCase userusecase.UserVerificationResendUseCase
	verifyUseCase             userusecase.UserVerifyUseCase
	loginUseCase              userusecase.UserLoginUseCase
	refreshUseCase            userusecase.UserTokenRefreshUseCase
	logoutUseCase             userusecase.UserLogoutUseCase
	deleteUseCase             userusecase.UserDeleteUseCase
	getUseCase                userusecase.UserGetUseCase
	listUseCase               userusecase.UserListUseCase
	updateUseCase             userusecase.UserUpdateUseCase
	changePasswordUseCase     userusecase.UserChangePasswordUseCase
}

func NewUserService(
	registerUseCase userusecase.UserRegisterUseCase,
	resendUseCase userusecase.UserVerificationResendUseCase,
	verifyUseCase userusecase.UserVerifyUseCase,
	loginUseCase userusecase.UserLoginUseCase,
	refreshUseCase userusecase.UserTokenRefreshUseCase,
	logoutUseCase userusecase.UserLogoutUseCase,
	deleteUseCase userusecase.UserDeleteUseCase,
	getUseCase userusecase.UserGetUseCase,
	listUseCase userusecase.UserListUseCase,
	updateUseCase userusecase.UserUpdateUseCase,
	changePasswordUseCase userusecase.UserChangePasswordUseCase,
) *UserService {
	return &UserService{
		registerUseCase:           registerUseCase,
		resendVerificationUseCase: resendUseCase,
		verifyUseCase:             verifyUseCase,
		loginUseCase:              loginUseCase,
		refreshUseCase:            refreshUseCase,
		logoutUseCase:             logoutUseCase,
		deleteUseCase:             deleteUseCase,
		getUseCase:                getUseCase,
		listUseCase:               listUseCase,
		updateUseCase:             updateUseCase,
		changePasswordUseCase:     changePasswordUseCase,
	}
}

func (s *UserService) RegisterService(server grpc.ServiceRegistrar) {
	user.RegisterUserServiceServer(server, s)
}

func (s *UserService) Register(ctx context.Context, req *user.RegisterRequest) (*user.RegisterResponse, error) {
	if err := s.registerUseCase.RegisterAccount(ctx, userusecase.RequestRegister{
		Username:         req.GetUsername(),
		Password:         req.GetPassword(),
		Email:            req.GetEmail(),
		Name:             req.GetName(),
		Cmnd:             req.GetCmnd(),
		Birthday:         req.GetBirthday(),
		Gender:           req.GetGender(),
		PermanentAddress: req.GetPermanentAddress(),
		PhoneNumber:      req.GetPhoneNumber(),
	}); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to register account: %v", err)
	}

	return &user.RegisterResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
	}, nil
}

func (s *UserService) ResendVerification(ctx context.Context, req *user.ResendVerificationRequest) (*user.ResendVerificationResponse, error) {
	if err := s.resendVerificationUseCase.Resend(ctx, userusecase.RequestResendVerification{Email: req.GetEmail()}); err != nil {
		switch {
		case errors.Is(err, userusecase.ErrResendEmptyEmail):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, userusecase.ErrResendAlreadyVerified):
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "resend verification: %v", err)
		}
	}

	return &user.ResendVerificationResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
	}, nil
}

func (s *UserService) VerifyUser(ctx context.Context, req *user.VerifyUserRequest) (*user.VerifyUserResponse, error) {
	entity, err := s.verifyUseCase.Verify(ctx, req.GetToken())
	if err != nil {
		switch {
		case errors.Is(err, userusecase.ErrVerifyEmptyToken):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, userusecase.ErrVerifyTokenExpired), errors.Is(err, userusecase.ErrVerifyTokenUsed):
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		default:
			return nil, status.Errorf(codes.Internal, "verify user: %v", err)
		}
	}

	return &user.VerifyUserResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
		Data:    toUserProfile(entity),
	}, nil
}

func (s *UserService) Login(ctx context.Context, req *user.LoginRequest) (*user.LoginResponse, error) {
	token, tokenExpire, refresh, refreshExpire, _, err := s.loginUseCase.Login(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "invalid credentials") {
			return nil, status.Errorf(codes.PermissionDenied, "login: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "login: %v", err)
	}

	return &user.LoginResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
		Data: &user.LoginResponse_Data{
			Token:                     token,
			TokenExpiresAtUnix:        tokenExpire.Unix(),
			RefreshToken:              refresh,
			RefreshTokenExpiresAtUnix: refreshExpire.Unix(),
		},
	}, nil
}

func (s *UserService) RefreshToken(ctx context.Context, req *user.RefreshTokenRequest) (*user.RefreshTokenResponse, error) {
	result, err := s.refreshUseCase.Refresh(ctx, req.GetRefreshToken())
	if err != nil {
		errLower := strings.ToLower(err.Error())
		if strings.Contains(errLower, "validate refresh token") || strings.Contains(errLower, "revoke refresh token") {
			return nil, status.Errorf(codes.PermissionDenied, "refresh token: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "refresh token: %v", err)
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

func (s *UserService) Logout(ctx context.Context, req *user.LogoutRequest) (*user.LogoutResponse, error) {
	if err := s.logoutUseCase.Logout(ctx, req.GetRefreshToken()); err != nil {
		return nil, status.Errorf(codes.Internal, "logout: %v", err)
	}

	return &user.LogoutResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
	}, nil
}

func (s *UserService) Delete(ctx context.Context, req *user.DeleteRequest) (*user.DeleteResponse, error) {
	if err := s.deleteUseCase.DeleteAccount(ctx, req.GetUsername()); err != nil {
		switch {
		case errors.Is(err, userusecase.ErrEmptyUsername):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, userusecase.ErrPermissionDenied):
			return nil, status.Error(codes.PermissionDenied, err.Error())
		default:
			if strings.Contains(strings.ToLower(err.Error()), "not found") {
				return nil, status.Errorf(codes.NotFound, "delete user: %v", err)
			}
			return nil, status.Errorf(codes.Internal, "delete user: %v", err)
		}
	}

	return &user.DeleteResponse{
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

	return &user.GetUserResponse{
		Code:    uint32(codes.OK),
		Message: codes.OK.String(),
		Data:    toUserProfile(entity),
	}, nil
}

func (s *UserService) List(ctx context.Context, req *user.ListUsersRequest) (*user.ListUsersResponse, error) {
	result, err := s.listUseCase.List(ctx, req.GetPage(), req.GetPageSize())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list users: %v", err)
	}

	profiles := make([]*user.UserProfile, 0, len(result.Users))
	for _, entity := range result.Users {
		profiles = append(profiles, toUserProfile(entity))
	}

	return &user.ListUsersResponse{
		Code:     uint32(codes.OK),
		Message:  codes.OK.String(),
		Data:     profiles,
		Total:    uint64(result.Total),
		Page:     result.Page,
		PageSize: result.PageSize,
	}, nil
}

func (s *UserService) Update(ctx context.Context, req *user.UpdateUserRequest) (*user.UpdateUserResponse, error) {
	err := s.updateUseCase.UpdateProfile(ctx, req.GetUsername(), userusecase.RequestUpdate{
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
		case errors.Is(err, userusecase.ErrUpdateEmptyUsername), errors.Is(err, userusecase.ErrUpdateEmptyEmail), errors.Is(err, userusecase.ErrUpdateEmptyName):
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
		case errors.Is(err, userusecase.ErrChangePasswordEmptyUsername), errors.Is(err, userusecase.ErrChangePasswordEmptyNew):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case errors.Is(err, userusecase.ErrChangePasswordInvalidCurrent):
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

func toUserProfile(entity userentity.User) *user.UserProfile {
	var (
		birthday   int64
		created    int64
		updated    int64
		verifiedAt int64
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
	if !entity.VerifiedAt.IsZero() {
		verifiedAt = entity.VerifiedAt.Unix()
	}

	return &user.UserProfile{
		Id:               entity.Id,
		Username:         entity.Username,
		Name:             entity.Name,
		Email:            entity.Email,
		Cmnd:             entity.DocumentID,
		Birthday:         birthday,
		Gender:           entity.Gender,
		PermanentAddress: entity.PermanentAddress,
		PhoneNumber:      entity.PhoneNumber,
		CreatedAt:        created,
		UpdatedAt:        updated,
		Verified:         entity.Verified,
		VerifiedAt:       verifiedAt,
	}
}
