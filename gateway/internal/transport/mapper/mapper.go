package transport_mapper

import (
	"uuid"

	core_model "github.com/GinciuuKitakaze/gateway/internal/core/model"
	"github.com/GinciuuKitakaze/gateway/internal/transport/dto"
)

// ToDomainRegister — преобразует DTO регистрации
// во внутреннюю модель Gateway.
func ToDomainRegister(req dto.RegisterRequest) core_model.Register {
	return core_model.Register{
		Name:     req.Name,
		Email:    req.Email,
		Phone:    req.Phone,
		Password: req.Password,
	}
}

// ToDTORegister — формирует HTTP-ответ после успешной регистрации.
func ToDTORegister(
	user core_model.User,
	tokenPair core_model.TokenPair,
) dto.RegisterResponse {
	return dto.RegisterResponse{
		User: dto.UserResponse{
			ID:    user.ID,
			Name:  user.Name,
			Email: user.Email,
			Phone: user.Phone,
		},
		TokenPair: dto.TokenPairResponse{
			AccessToken:  tokenPair.AccessToken,
			RefreshToken: tokenPair.RefreshToken,
		},
	}
}

// ToDomainLogin — преобразует DTO авторизации
// во внутреннюю модель Gateway.
func ToDomainLogin(req dto.LoginRequest) core_model.Login {
	return core_model.Login{
		EmailOrPhone: req.EmailOrPhone,
		Password:     req.Password,
	}
}

// ToDTOTokenPair — преобразует внутреннюю пару токенов
// в HTTP DTO ответа.
func ToDTOTokenPair(
	tokenPair core_model.TokenPair,
) dto.TokenPairResponse {
	return dto.TokenPairResponse{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}
}

// ToDomainAuthLogout — преобразует DTO выхода
// во внутреннюю модель AuthLogout.
func ToDomainAuthLogout(
	req dto.AuthLogoutRequest,
) core_model.AuthLogout {
	return core_model.AuthLogout{
		AccessToken:  req.AccessToken,
		RefreshToken: req.RefreshToken,
	}
}

// ToDomainUser — преобразует DTO пользователя
// во внутреннюю модель Gateway.
func ToDomainUser(
	user dto.UserRequest,
) core_model.User {
	return core_model.User{
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
	}
}

// ToDTOUser — преобразует внутреннюю модель пользователя
// в HTTP DTO ответа.
func ToDTOUser(
	user core_model.User,
) dto.UserResponse {
	return dto.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
		Phone: user.Phone,
	}
}

// ToDTOUsers — преобразует список пользователей
// в список HTTP DTO.
func ToDTOUsers(
	users []core_model.User,
) []dto.UserResponse {
	res := make([]dto.UserResponse, len(users))

	for i, user := range users {
		res[i] = ToDTOUser(user)
	}

	return res
}

// ToDomainUpdateUser — объединяет ID пользователя из path
// и изменяемые поля из HTTP body.
func ToDomainUpdateUser(
	userID uuid.UUID,
	req dto.UpdateUserRequest,
) core_model.UpdateUser {
	return core_model.UpdateUser{
		ID:    userID,
		Name:  req.Name,
		Email: req.Email,
		Phone: req.Phone,
	}
}
