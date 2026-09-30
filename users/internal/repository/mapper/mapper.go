package repository_mapper

import (
	core_model "github.com/GinciuuKitakaze/users/internal/core/model"
	repository_model "github.com/GinciuuKitakaze/users/internal/repository/model"
)

func ToDomain(user repository_model.User) core_model.User {
	return core_model.User{
		ID:        user.ID,
		Version:   user.Version,
		Name:      user.Name,
		Email:     user.Email,
		Phone:     user.Phone,
		IsDeleted: user.IsDeleted,
		Created:   user.Created,
		Updated:   user.Updated,
	}
}

func ToDomains(user []repository_model.User) []core_model.User {
	models := make([]core_model.User, len(user))
	for i, user := range user {
		models[i] = ToDomain(user)
	}
	return models
}
