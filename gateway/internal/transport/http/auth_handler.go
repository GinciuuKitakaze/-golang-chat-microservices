package httptransport

import (
	"net/http"

	"github.com/GinciuuKitakaze/gateway/internal/transport/dto"
	http_response "github.com/GinciuuKitakaze/gateway/internal/transport/http/response"
	transport_mapper "github.com/GinciuuKitakaze/gateway/internal/transport/mapper"
	"github.com/GinciuuKitakaze/gateway/internal/validator"
)

// Register — регистрирует нового пользователя через Gateway.
func (s *Server) Register(
	rw http.ResponseWriter,
	r *http.Request,
) {
	responseHandler := http_response.NewHTTPResponseHandler(
		s.logger,
		rw,
	)

	var req dto.RegisterRequest

	// Читаем JSON body запроса.
	if err := validator.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ClientErrorResponse(
			err,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	register := transport_mapper.ToDomainRegister(req)

	// Создаём пользователя и получаем пару токенов.
	user, tokenPair, err := s.gatewayService.Register(
		r.Context(),
		register,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"register failed",
		)
		return
	}

	responseHandler.JSONResponse(
		transport_mapper.ToDTORegister(
			user,
			tokenPair,
		),
		http.StatusCreated,
	)
}

// Login — авторизует пользователя по email или телефону.
func (s *Server) Login(
	rw http.ResponseWriter,
	r *http.Request,
) {
	responseHandler := http_response.NewHTTPResponseHandler(
		s.logger,
		rw,
	)

	var req dto.LoginRequest

	// Получаем email/phone и пароль из body.
	if err := validator.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ClientErrorResponse(
			err,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	login := transport_mapper.ToDomainLogin(req)

	tokenPair, err := s.gatewayService.Login(
		r.Context(),
		login,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"login failed",
		)
		return
	}

	responseHandler.JSONResponse(
		transport_mapper.ToDTOTokenPair(tokenPair),
		http.StatusOK,
	)
}

// Logout — завершает пользовательскую сессию.
func (s *Server) Logout(
	rw http.ResponseWriter,
	r *http.Request,
) {
	responseHandler := http_response.NewHTTPResponseHandler(
		s.logger,
		rw,
	)

	var req dto.AuthLogoutRequest

	// Получаем access и refresh токены из body.
	if err := validator.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ClientErrorResponse(
			err,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	authLogout := transport_mapper.ToDomainAuthLogout(req)

	if err := s.gatewayService.Logout(
		r.Context(),
		authLogout,
	); err != nil {
		responseHandler.ErrorResponse(
			err,
			"logout failed",
		)
		return
	}

	responseHandler.NoContentResponse()
}

// Refresh — обновляет пару access/refresh токенов.
func (s *Server) Refresh(
	rw http.ResponseWriter,
	r *http.Request,
) {
	responseHandler := http_response.NewHTTPResponseHandler(
		s.logger,
		rw,
	)

	var req dto.RefreshTokenRequest

	// Получаем refresh token из body.
	if err := validator.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ClientErrorResponse(
			err,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	tokenPair, err := s.gatewayService.Refresh(
		r.Context(),
		req.RefreshToken,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"refresh token failed",
		)
		return
	}

	responseHandler.JSONResponse(
		transport_mapper.ToDTOTokenPair(tokenPair),
		http.StatusOK,
	)
}

// ValidateToken — проверяет access token.
//
// Позже этот вызов будет использоваться внутри auth middleware.
func (s *Server) ValidateToken(
	rw http.ResponseWriter,
	r *http.Request,
) {
	responseHandler := http_response.NewHTTPResponseHandler(
		s.logger,
		rw,
	)

	var req dto.AccessTokenRequest

	// Получаем access token из body.
	if err := validator.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ClientErrorResponse(
			err,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	userID, err := s.gatewayService.ValidateToken(
		r.Context(),
		req.AccessToken,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"validate token failed",
		)
		return
	}

	responseHandler.JSONResponse(
		userID,
		http.StatusOK,
	)
}
