package users

import (
	"context"
	"fmt"
	"uuid"

	core_model "github.com/GinciuuKitakaze/gateway/internal/core/model"
	userspb "github.com/GinciuuKitakaze/messenger-contracts/users/go"
)

type Service struct {
	client userspb.UserServiceClient
}

func NewService(client userspb.UserServiceClient) *Service {
	return &Service{
		client: client,
	}
}

// CreateUser — создание пользователя в Users Service.
func (s *Service) CreateUser(
	ctx context.Context,
	req core_model.User,
) (core_model.User, error) {

	res, err := s.client.CreateUser(ctx, &userspb.CreateUserRequest{
		User: &userspb.CreateUser{
			Name:  req.Name,
			Email: req.Email,
			Phone: req.Phone,
		},
	})
	if err != nil {
		return core_model.User{}, fmt.Errorf(
			"create user: %w",
			err,
		)
	}

	user, err := toDomainUser(res.GetUser())
	if err != nil {
		return core_model.User{}, fmt.Errorf(
			"map created user: %w",
			err,
		)
	}

	return user, nil
}

// GetUser — получение пользователя по ID.
func (s *Service) GetUser(
	ctx context.Context,
	userID uuid.UUID,
) (core_model.User, error) {

	res, err := s.client.GetUser(ctx, &userspb.GetUserRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return core_model.User{}, fmt.Errorf(
			"get user: %w",
			err,
		)
	}

	user, err := toDomainUser(res.GetUser())
	if err != nil {
		return core_model.User{}, fmt.Errorf(
			"map user: %w",
			err,
		)
	}

	return user, nil
}

// GetUsers — получение списка пользователей.
func (s *Service) GetUsers(
	ctx context.Context,
	limit,
	offset uint32,
) ([]core_model.User, error) {

	res, err := s.client.GetUsers(ctx, &userspb.GetUsersRequest{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf(
			"get users: %w",
			err,
		)
	}

	users := res.GetUsers()

	userList := make([]core_model.User, 0, len(users))

	for _, user := range users {
		domainUser, err := toDomainUser(user)
		if err != nil {
			return nil, fmt.Errorf(
				"map user: %w",
				err,
			)
		}

		userList = append(userList, domainUser)
	}

	return userList, nil
}

// GetUserByEmailOrPhone — получение пользователя по email или телефону.
func (s *Service) GetUserByEmailOrPhone(
	ctx context.Context,
	emailOrPhone string,
) (core_model.User, error) {

	res, err := s.client.GetUserByEmailOrPhone(
		ctx,
		&userspb.GetUserByEmailOrPhoneRequest{
			EmailOrPhone: emailOrPhone,
		},
	)
	if err != nil {
		return core_model.User{}, fmt.Errorf(
			"get user by email or phone: %w",
			err,
		)
	}

	user, err := toDomainUser(res.GetUser())
	if err != nil {
		return core_model.User{}, fmt.Errorf(
			"map user: %w",
			err,
		)
	}

	return user, nil
}

// DeleteUser — удаление пользователя.
func (s *Service) DeleteUser(
	ctx context.Context,
	userID uuid.UUID,
) (core_model.User, error) {

	res, err := s.client.DeleteUser(ctx, &userspb.DeleteUserRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return core_model.User{}, fmt.Errorf(
			"delete user: %w",
			err,
		)
	}

	user, err := toDomainUser(res.GetUser())
	if err != nil {
		return core_model.User{}, fmt.Errorf(
			"map deleted user: %w",
			err,
		)
	}

	return user, nil
}

// UpdateUser — обновление пользователя.
func (s *Service) UpdateUser(
	ctx context.Context,
	updatedUser core_model.UpdateUser,
) (core_model.User, error) {

	res, err := s.client.UpdateUser(ctx, &userspb.UpdateUserRequest{
		UserId: updatedUser.ID.String(),
		User: &userspb.UpdateUser{
			Name:  updatedUser.Name,
			Email: updatedUser.Email,
			Phone: updatedUser.Phone,
		},
	})
	if err != nil {
		return core_model.User{}, fmt.Errorf(
			"update user: %w",
			err,
		)
	}

	user, err := toDomainUser(res.GetUser())
	if err != nil {
		return core_model.User{}, fmt.Errorf(
			"map updated user: %w",
			err,
		)
	}

	return user, nil
}

// toDomainUser — преобразование protobuf-модели пользователя
// во внутреннюю модель Gateway.
func toDomainUser(user *userspb.User) (core_model.User, error) {
	if user == nil {
		return core_model.User{}, fmt.Errorf("user is nil")
	}

	userID, err := uuid.Parse(user.GetId())
	if err != nil {
		return core_model.User{}, fmt.Errorf(
			"parse user id: %w",
			err,
		)
	}

	return core_model.User{
		ID:    userID,
		Name:  user.GetName(),
		Email: user.Email,
		Phone: user.Phone,
	}, nil
}
