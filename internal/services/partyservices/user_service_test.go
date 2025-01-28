package partyservices

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	commonproto "github.com/prov100/dc2/internal/protogen/common/v1"
	partyproto "github.com/prov100/dc2/internal/protogen/party/v1"
	"github.com/prov100/dc2/test"
	"github.com/stretchr/testify/assert"
)

func TestUserService_GetUsers(t *testing.T) {
	err := test.LoadSQL(logUser, dbService)
	fmt.Println("TestUserService_GetUsers err", err)
	if err != nil {
		t.Error(err)
		return
	}

	ctx := LoginUser()
	// fmt.Println("TestUserService_GetUsers serverOpt", serverOpt)
	// fmt.Println("TestUserService_GetUsers userOpt", userOpt)
	// fmt.Println("TestUserService_GetUsers redisService", redisService)
	userService := NewUserService(log, dbService, redisService, mailerService, jwtOpt, userOpt, serverOpt)

	fmt.Println("TestUserService_GetUsers userService", userService)

	user1, err := GetUser("auth0|66fd06d0bfea78a82bb42459", "sprov300@gmail.com", "https://s.gravatar.com/avatar/52ab1cc37bb42deb67ea939fd68ff7d4?s=480&r=pg&d=https%3A%2F%2Fcdn.auth0.com%2Favatars%2Fsp.png", "sprov300@gmail.com")
	if err != nil {
		t.Error(err)
		return
	}

	user2, err := GetUser("auth0|66fcdfb6d20dcb68e3fcbc3b", "sprov200@gmail.com", "https://s.gravatar.com/avatar/06004bcbe9705b0ba5d7c4923fef0061?s=480&r=pg&d=https%3A%2F%2Fcdn.auth0.com%2Favatars%2Fsp.png", "sprov200@gmail.com")
	if err != nil {
		t.Error(err)
		return
	}

	users := []*partyproto.User{}
	users = append(users, user1)
	users = append(users, user2)

	usersResponse := partyproto.GetUsersResponse{}
	usersResponse.Users = users

	form := partyproto.GetUsersRequest{}
	form.UserEmail = "sprov300@gmail.com"
	form.RequestId = "bks1m1g91jau4nkks2f0"

	type args struct {
		ctx context.Context
		in  *partyproto.GetUsersRequest
	}
	tests := []struct {
		u       *UserService
		args    args
		want    *partyproto.GetUsersResponse
		wantErr bool
	}{
		{
			u: userService,
			args: args{
				ctx: ctx,
				in:  &form,
			},
			want:    &usersResponse,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		// fmt.Println("TestUserService_GetUsers tt", tt)
		// fmt.Println("TestUserService_GetUsers tt.u", tt.u)
		// fmt.Println("TestUserService_GetUsers tt.args.ctx", tt.args.ctx)
		// fmt.Println("TestUserService_GetUsers tt.args.in", tt.args.in)
		usersResp, err := tt.u.GetUsers(tt.args.ctx, tt.args.in)
		fmt.Println("TestUserService_GetUsers usersResp", usersResp)
		fmt.Println("TestUserService_GetUsers usersResp err", err)
		if (err != nil) != tt.wantErr {
			t.Errorf("UserService.GetUsers() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		/*if !reflect.DeepEqual(usersResp, tt.want) {
			t.Errorf("UserService.GetUsers() = %v want %v", usersResp, tt.want)
		}*/

		assert.NotNil(t, usersResp)
		userResult := usersResp.Users[0]
		assert.Equal(t, userResult.Name, "sprov300@gmail.com", "they should be equal")
		assert.Equal(t, userResult.Picture, "https://s.gravatar.com/avatar/52ab1cc37bb42deb67ea939fd68ff7d4?s=480&r=pg&d=https%3A%2F%2Fcdn.auth0.com%2Favatars%2Fsp.png", "they should be equal")
		assert.Equal(t, userResult.Email, "sprov300@gmail.com", "they should be equal")
	}
}

func TestUserService_GetUserByEmail(t *testing.T) {
	err := test.LoadSQL(logUser, dbService)
	if err != nil {
		t.Error(err)
		return
	}

	ctx := LoginUser()

	userService := NewUserService(log, dbService, redisService, mailerService, jwtOpt, userOpt, serverOpt)

	user, err := GetUser("auth0|66fd06d0bfea78a82bb42459", "sprov300@gmail.com", "https://s.gravatar.com/avatar/52ab1cc37bb42deb67ea939fd68ff7d4?s=480&r=pg&d=https%3A%2F%2Fcdn.auth0.com%2Favatars%2Fsp.png", "sprov300@gmail.com")
	if err != nil {
		t.Error(err)
		return
	}

	userResponse := partyproto.GetUserByEmailResponse{}
	userResponse.User = user

	form := partyproto.GetUserByEmailRequest{}
	form.Email = "sprov300@gmail.com"
	form.UserEmail = "sprov300@gmail.com"
	form.RequestId = "bks1m1g91jau4nkks2f0"

	type args struct {
		ctx context.Context
		in  *partyproto.GetUserByEmailRequest
	}
	tests := []struct {
		u       *UserService
		args    args
		want    *partyproto.GetUserByEmailResponse
		wantErr bool
	}{
		{
			u: userService,
			args: args{
				ctx: ctx,
				in:  &form,
			},
			want:    &userResponse,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		userResp, err := tt.u.GetUserByEmail(tt.args.ctx, tt.args.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("UserService.GetUserByEmail() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(userResp, tt.want) {
			t.Errorf("UserService.GetUserByEmail() = %v, want %v", userResp, tt.want)
		}
		assert.NotNil(t, userResp)
		userResult := userResp.User
		assert.Equal(t, userResult.Name, "sprov300@gmail.com", "they should be equal")
		assert.Equal(t, userResult.Picture, "https://s.gravatar.com/avatar/52ab1cc37bb42deb67ea939fd68ff7d4?s=480&r=pg&d=https%3A%2F%2Fcdn.auth0.com%2Favatars%2Fsp.png", "they should be equal")
		assert.Equal(t, userResult.Email, "sprov300@gmail.com", "they should be equal")
	}
}

func TestUserService_GetUser(t *testing.T) {
	err := test.LoadSQL(logUser, dbService)
	if err != nil {
		t.Error(err)
		return
	}

	ctx := LoginUser()

	userService := NewUserService(log, dbService, redisService, mailerService, jwtOpt, userOpt, serverOpt)

	user, err := GetUser("auth0|66fd06d0bfea78a82bb42459", "sprov300@gmail.com", "https://s.gravatar.com/avatar/52ab1cc37bb42deb67ea939fd68ff7d4?s=480&r=pg&d=https%3A%2F%2Fcdn.auth0.com%2Favatars%2Fsp.png", "sprov300@gmail.com")
	if err != nil {
		t.Error(err)
		return
	}

	userResponse := partyproto.GetUserResponse{}
	userResponse.User = user

	form := partyproto.GetUserRequest{}
	gform := commonproto.GetRequest{}
	gform.Id = "auth0|66fd06d0bfea78a82bb42459"
	gform.UserEmail = "sprov300@gmail.com"
	gform.RequestId = "bks1m1g91jau4nkks2f0"
	form.GetRequest = &gform

	type args struct {
		ctx context.Context
		in  *partyproto.GetUserRequest
	}
	tests := []struct {
		u       *UserService
		args    args
		want    *partyproto.GetUserResponse
		wantErr bool
	}{
		{
			u: userService,
			args: args{
				ctx: ctx,
				in:  &form,
			},
			want:    &userResponse,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		userResp, err := tt.u.GetUser(tt.args.ctx, tt.args.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("UserService.GetUser() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(userResp, tt.want) {
			t.Errorf("UserService.GetUser() = %v, want %v", userResp, tt.want)
		}
		assert.NotNil(t, userResp)
		userResult := userResp.User
		assert.Equal(t, userResult.Name, "sprov300@gmail.com", "they should be equal")
		assert.Equal(t, userResult.Picture, "https://s.gravatar.com/avatar/52ab1cc37bb42deb67ea939fd68ff7d4?s=480&r=pg&d=https%3A%2F%2Fcdn.auth0.com%2Favatars%2Fsp.png", "they should be equal")
		assert.Equal(t, userResult.Email, "sprov300@gmail.com", "they should be equal")
	}
}

func GetUser(id string, email string, picture string, name string) (*partyproto.User, error) {
	user := new(partyproto.User)
	user.Id = id
	user.Email = email
	user.Picture = picture
	user.Name = name
	return user, nil
}

/*func TestUserService_UpdateUser(t *testing.T) {
	err := test.LoadSQL(logUser, dbService)
	if err != nil {
		t.Error(err)
		return
	}

	ctx := LoginUser()

	userService := NewUserService(log, dbService, redisService, tokenService, mailerService, jwtOpt, userOpt)

	form1 := partyproto.UpdateUserRequest{}
	form1.Id = "ca2862cc-97ae-4705-b372-de87762b22f6"
	form1.FirstName = "PsskZoQ"
	form1.LastName = "Distributor1"
	form1.UserId = "ca2862cc-97ae-4705-b372-de87762b22f6"
	form1.UserEmail = "sprov100@gmail.com"
	form1.RequestId = "bks1m1g91jau4nkks2f0"

	partyResponse := partyproto.UpdateUserResponse{}

	type args struct {
		ctx context.Context
		in  *partyproto.UpdateUserRequest
	}
	tests := []struct {
		u       *UserService
		args    args
		want    *partyproto.UpdateUserResponse
		wantErr bool
	}{
		{
			u: userService,
			args: args{
				ctx: ctx,
				in:  &form1,
			},
			want:    &partyResponse,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := tt.u.UpdateUser(tt.args.ctx, tt.args.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("UserService.UpdateUser() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("UserService.UpdateUser() = %v, want %v", got, tt.want)
		}
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	err := test.LoadSQL(logUser, dbService)
	if err != nil {
		t.Error(err)
		return
	}

	ctx := LoginUser()

	userService := NewUserService(log, dbService, redisService, tokenService, mailerService, jwtOpt, userOpt)

	form := partyproto.DeleteUserRequest{}
	gform := commonproto.GetRequest{}
	gform.Id = "ca2862cc-97ae-4705-b372-de87762b22f6"
	gform.UserEmail = "sprov100@gmail.com"
	gform.RequestId = "bks1m1g91jau4nkks2f0"
	form.GetRequest = &gform

	partyResponse := partyproto.DeleteUserResponse{}

	type args struct {
		ctx context.Context
		in  *partyproto.DeleteUserRequest
	}
	tests := []struct {
		u       *UserService
		args    args
		want    *partyproto.DeleteUserResponse
		wantErr bool
	}{
		{
			u: userService,
			args: args{
				ctx: ctx,
				in:  &form,
			},
			want:    &partyResponse,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := tt.u.DeleteUser(tt.args.ctx, tt.args.in)
		if (err != nil) != tt.wantErr {
			t.Errorf("UserService.DeleteUser() error = %v, wantErr %v", err, tt.wantErr)
			return
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("UserService.DeleteUser() = %v, want %v", got, tt.want)
		}
	}
}*/
