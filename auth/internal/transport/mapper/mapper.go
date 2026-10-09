package transport_mapper

import (
	"uuid"

	core_model "github.com/GinciuuKitakaze/auth/internal/core/model"
	authpb "github.com/GinciuuKitakaze/messenger-contracts/auth/go"
)

func ToDomainRegister(register *authpb.RegisterRequest, userID uuid.UUID) core_model.Register {
	return core_model.Register{
		UserID:   userID,
		Password: register.Password,
	}
}

func ToDomainLogin(login *authpb.LoginRequest, userID uuid.UUID) core_model.Login {
	return core_model.Login{
		UserID:   userID,
		Password: login.Password,
	}
}

func ToProtoTokenPair(tokenPair core_model.TokenPair) *authpb.TokenPair {
	return &authpb.TokenPair{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}
}

func ToProtoLogin(tokenPair core_model.TokenPair) *authpb.LoginResponse {
	return &authpb.LoginResponse{
		TokenPair: ToProtoTokenPair(tokenPair),
	}
}

func ToProtoRefresh(tokenPair core_model.TokenPair) *authpb.RefreshResponse {
	return &authpb.RefreshResponse{
		TokenPair: ToProtoTokenPair(tokenPair),
	}
}

func ToProtoValidate(userID string) *authpb.ValidateResponse {
	return &authpb.ValidateResponse{
		UserId: userID,
	}
}
