package transport_mapper

import (
	accountpb "github.com/GinciuuKitakaze/messenger-contracts/users/go"
	core_model "github.com/GinciuuKitakaze/users/internal/core/model"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToProtoUser(u core_model.User) *accountpb.User {
	return &accountpb.User{
		Id:        u.ID.String(),
		Version:   u.Version,
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone,
		CreatedAt: timestamppb.New(u.Created),
		UpdatedAt: timestamppb.New(u.Updated),
	}
}

func ToProtoUsers(users []core_model.User) []*accountpb.User {
	pbs := make([]*accountpb.User, len(users))
	for i, u := range users {
		pbs[i] = ToProtoUser(u)
	}
	return pbs
}

func ToModelCreateUser(u *accountpb.CreateUser) core_model.CreateUser {
	return core_model.CreateUser{
		Name:  u.Name,
		Email: u.Email,
		Phone: u.Phone,
	}
}

func ToModelUpdateUser(u *accountpb.UpdateUser) core_model.UpdateUser {
	return core_model.UpdateUser{
		Name:  u.Name,
		Email: u.Email,
		Phone: u.Phone,
	}
}
