package httptransport

import (
	"net/http"
	"strconv"
	"uuid"

	"github.com/GinciuuKitakaze/gateway/internal/transport/dto"
	http_response "github.com/GinciuuKitakaze/gateway/internal/transport/http/response"
	transport_mapper "github.com/GinciuuKitakaze/gateway/internal/transport/mapper"
	"github.com/GinciuuKitakaze/gateway/internal/validator"
)

// CreateUser — создаёт пользователя напрямую через Users Service.
//
// Публичный route для этого handler пока не регистрируем,
// чтобы нельзя было создать пользователя без Auth credentials.
func (s *Server) CreateUser(
	rw http.ResponseWriter,
	r *http.Request,
) {
	responseHandler := http_response.NewHTTPResponseHandler(
		s.logger,
		rw,
	)

	var req dto.UserRequest

	// Читаем данные пользователя из body.
	if err := validator.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ClientErrorResponse(
			err,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	userDTO := transport_mapper.ToDomainUser(req)

	user, err := s.gatewayService.CreateUser(
		r.Context(),
		userDTO,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"create user failed",
		)
		return
	}

	responseHandler.JSONResponse(
		transport_mapper.ToDTOUser(user),
		http.StatusCreated,
	)
}

// GetUser — возвращает пользователя по ID.
func (s *Server) GetUser(
	rw http.ResponseWriter,
	r *http.Request,
) {
	responseHandler := http_response.NewHTTPResponseHandler(
		s.logger,
		rw,
	)

	// ID пользователя получаем из path:
	// /api/v1/users/{id}
	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		responseHandler.ClientErrorResponse(
			err,
			"invalid user id",
			http.StatusBadRequest,
		)
		return
	}

	user, err := s.gatewayService.GetUser(
		r.Context(),
		userID,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"get user failed",
		)
		return
	}

	responseHandler.JSONResponse(
		transport_mapper.ToDTOUser(user),
		http.StatusOK,
	)
}

// GetUsers — возвращает список пользователей с пагинацией.
func (s *Server) GetUsers(
	rw http.ResponseWriter,
	r *http.Request,
) {
	responseHandler := http_response.NewHTTPResponseHandler(
		s.logger,
		rw,
	)

	// Значения пагинации по умолчанию.
	limit := uint32(20)
	offset := uint32(0)

	// limit передаётся через query:
	// /api/v1/users?limit=20
	if value := r.URL.Query().Get("limit"); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			responseHandler.ClientErrorResponse(
				err,
				"invalid limit",
				http.StatusBadRequest,
			)
			return
		}

		limit = uint32(parsed)
	}

	if limit == 0 {
		limit = 20
	}

	if limit > 100 {
		limit = 100
	}

	// offset также передаётся через query.
	if value := r.URL.Query().Get("offset"); value != "" {
		parsed, err := strconv.ParseUint(value, 10, 32)
		if err != nil {
			responseHandler.ClientErrorResponse(
				err,
				"invalid offset",
				http.StatusBadRequest,
			)
			return
		}

		offset = uint32(parsed)
	}

	users, err := s.gatewayService.GetUsers(
		r.Context(),
		limit,
		offset,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"get users failed",
		)
		return
	}

	responseHandler.JSONResponse(
		transport_mapper.ToDTOUsers(users),
		http.StatusOK,
	)
}

// DeleteUser — полностью удаляет пользователя.
func (s *Server) DeleteUser(
	rw http.ResponseWriter,
	r *http.Request,
) {
	responseHandler := http_response.NewHTTPResponseHandler(
		s.logger,
		rw,
	)

	// ID пользователя получаем из path.
	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		responseHandler.ClientErrorResponse(
			err,
			"invalid user id",
			http.StatusBadRequest,
		)
		return
	}

	user, err := s.gatewayService.DeleteUser(
		r.Context(),
		userID,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"delete user failed",
		)
		return
	}

	responseHandler.JSONResponse(
		transport_mapper.ToDTOUser(user),
		http.StatusOK,
	)
}

// UpdateUser — обновляет данные пользователя.
func (s *Server) UpdateUser(
	rw http.ResponseWriter,
	r *http.Request,
) {
	responseHandler := http_response.NewHTTPResponseHandler(
		s.logger,
		rw,
	)

	// ID пользователя получаем из path.
	userID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		responseHandler.ClientErrorResponse(
			err,
			"invalid user id",
			http.StatusBadRequest,
		)
		return
	}

	var req dto.UpdateUserRequest

	// Из body получаем только изменяемые поля.
	if err := validator.DecodeAndValidateRequest(r, &req); err != nil {
		responseHandler.ClientErrorResponse(
			err,
			"invalid request body",
			http.StatusBadRequest,
		)
		return
	}

	updateUser := transport_mapper.ToDomainUpdateUser(
		userID,
		req,
	)

	user, err := s.gatewayService.UpdateUser(
		r.Context(),
		updateUser,
	)
	if err != nil {
		responseHandler.ErrorResponse(
			err,
			"update user failed",
		)
		return
	}

	responseHandler.JSONResponse(
		transport_mapper.ToDTOUser(user),
		http.StatusOK,
	)
}
