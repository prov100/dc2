package partyservices

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/prov100/dc2/internal/common"
	"github.com/prov100/dc2/internal/config"
	commonproto "github.com/prov100/dc2/internal/protogen/common/v1"
	partyproto "github.com/prov100/dc2/internal/protogen/party/v1"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/contrib/propagators/aws/xray"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"

	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

/* error message range: 1500-1999 */

// UserService - For accessing user services
type UserService struct {
	log           *zap.Logger
	DBService     *common.DBService
	RedisService  *common.RedisService
	MailerService common.MailerIntf
	JWTOptions    *config.JWTOptions
	UserOptions   *config.UserOptions
	ServerOptions *config.ServerOptions
	partyproto.UnimplementedUserServiceServer
}

// NewUserService - Create User Service
func NewUserService(log *zap.Logger, dbOpt *common.DBService, redisOpt *common.RedisService, mailerOpt common.MailerIntf, jwtOptions *config.JWTOptions, userOpt *config.UserOptions, serverOpt *config.ServerOptions) *UserService {
	return &UserService{
		log:           log,
		DBService:     dbOpt,
		RedisService:  redisOpt,
		MailerService: mailerOpt,
		JWTOptions:    jwtOptions,
		UserOptions:   userOpt,
		ServerOptions: serverOpt,
	}
}

// Roles - Used for roles
type Roles []string

var log *zap.Logger

// StartUserServer - Start User Server
func StartUserServer(log *zap.Logger, isTest bool, pwd string, dbOpt *config.DBOptions, redisOpt *config.RedisOptions, mailerOpt *config.MailerOptions, serverOpt *config.ServerOptions, grpcServerOpt *config.GrpcServerOptions, jwtOpt *config.JWTOptions, oauthOpt *config.OauthOptions, userOpt *config.UserOptions, uptraceOpt *config.UptraceOptions, dbService *common.DBService, redisService *common.RedisService, mailerService common.MailerIntf) {
	common.SetJWTOpt(jwtOpt)
	// fmt.Println("grpcServerOpt is")
	creds, err := common.GetSrvCred(log, isTest, pwd, grpcServerOpt)
	// fmt.Println("creds")
	fmt.Println("creds err", err)
	if err != nil {
		os.Exit(1)
	}

	// tracer, _ := interceptors.NewJaegerTracer(log, jaegerTracerOpt, jaegerTracerOpt.UserServiceName)

	var srvOpts []grpc.ServerOption
	/*srvOpts = append(srvOpts, grpc.Creds(creds),
		grpc.ChainUnaryInterceptor(
			grpc_zap.UnaryServerInterceptor(log),
			interceptors.AccessLogUnaryServerInterceptor(),
			grpc_opentracing.UnaryServerInterceptor(grpc_opentracing.WithTracer(tracer)),
			interceptors.TokenAuthInterceptor,
			grpc.UnaryServerInterceptor(grpc_prometheus.UnaryServerInterceptor),
		),
	)*/

	srvOpts = append(srvOpts, grpc.Creds(creds))

	srvOpts = append(srvOpts, grpc.StatsHandler(otelgrpc.NewServerHandler()))

	userService := NewUserService(log, dbService, redisService, mailerService, jwtOpt, userOpt, serverOpt)
	fmt.Println("userservice StartUserServer userService", userService)
	lis, err := net.Listen("tcp", grpcServerOpt.GrpcUserServerPort)
	fmt.Println("userservice StartUserServer lis err", err)
	if err != nil {
		log.Error("Error", zap.Int("msgnum", 9109), zap.Error(err))
		os.Exit(1)
	}

	// Create a HTTP server for prometheus.
	// httpServer := &http.Server{Handler: promhttp.Handler(), Addr: promOpt.PromHTTPUserServerPort}

	srv := grpc.NewServer(srvOpts...)
	partyproto.RegisterUserServiceServer(srv, userService)

	// grpc_prometheus.Register(srv)

	// Start your http server for prometheus.
	/*go func() {
		if err := httpServer.ListenAndServe(); err != nil {
			fmt.Println("userservice StartUserServer err ListenAndServe", err)
			log.Error("Error", zap.Int("msgnum", 9112), zap.Error(err))
			os.Exit(1)
		}
	}()*/

	if err := srv.Serve(lis); err != nil {
		fmt.Println("userservice StartUserServer err srv.Serve", err)
		log.Error("Error", zap.Int("msgnum", 9414), zap.Error(err))
		os.Exit(1)
	}
}

// GetUsers - Get users
func (u *UserService) GetUsers(ctx context.Context, in *partyproto.GetUsersRequest) (*partyproto.GetUsersResponse, error) {
	// url := "https://dev-q3i1a0rds2l1nrpb.us.auth0.com/api/v2/users
	// fmt.Println("u.ServerOptions.Auth0Domain", u.ServerOptions.Auth0Domain)
	fmt.Println("u.ServerOptions.Auth0Domain", u.ServerOptions.Auth0MgmtToken)
	url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/users"
	respBody, err := common.SendRequest("GET", url, nil, "Bearer "+u.ServerOptions.Auth0MgmtToken)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}

	jsonDataReader := strings.NewReader(string(respBody))
	decoder := json.NewDecoder(jsonDataReader)
	var userResp []map[string]interface{}
	err = decoder.Decode(&userResp)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}
	users := []*partyproto.User{}
	for _, userResult := range userResp {
		user := partyproto.User{}
		user.Id = userResult["user_id"].(string)
		user.Email = userResult["email"].(string)
		user.Picture = userResult["picture"].(string)
		user.Name = userResult["name"].(string)
		users = append(users, &user)
	}

	usersResponse := partyproto.GetUsersResponse{}
	usersResponse.Users = users
	return &usersResponse, nil
}

// GetUserByEmail - Get user details by email
func (u *UserService) GetUserByEmail(ctx context.Context, in *partyproto.GetUserByEmailRequest) (*partyproto.GetUserByEmailResponse, error) {
	// url := "https://dev-q3i1a0rds2l1nrpb.us.auth0.com/api/v2/users-by-email?email=" + email
	fmt.Println("in.Email", in.Email)
	fmt.Println("u.ServerOptions.Auth0MgmtToken", "Bearer "+u.ServerOptions.Auth0MgmtToken)
	url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/users-by-email?email=" + in.Email
	respBody, err := common.SendRequest("GET", url, nil, "Bearer "+u.ServerOptions.Auth0MgmtToken)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}

	fmt.Println("respBody", respBody)

	jsonDataReader := strings.NewReader(string(respBody))
	decoder := json.NewDecoder(jsonDataReader)
	var userResp []map[string]interface{}
	err = decoder.Decode(&userResp)
	fmt.Println("userResp", userResp)
	fmt.Println("userResp", userResp[0])
	userResult := userResp[0]
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}

	user := partyproto.User{}
	user.Id = userResult["user_id"].(string)
	user.Email = userResult["email"].(string)
	user.Picture = userResult["picture"].(string)
	user.Name = userResult["name"].(string)

	userResponse := partyproto.GetUserByEmailResponse{}
	userResponse.User = &user
	return &userResponse, nil
}

// GetUser - used to get user by Id
func (u *UserService) GetUser(ctx context.Context, inReq *partyproto.GetUserRequest) (*partyproto.GetUserResponse, error) {
	in := inReq.GetRequest
	// url := "https://dev-q3i1a0rds2l1nrpb.us.auth0.com/api/v2/users/" + userID
	url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/users/" + in.Id
	respBody, err := common.SendRequest("GET", url, nil, "Bearer "+u.ServerOptions.Auth0MgmtToken)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}

	jsonDataReader := strings.NewReader(string(respBody))
	decoder := json.NewDecoder(jsonDataReader)
	var userResp map[string]interface{}
	err = decoder.Decode(&userResp)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}

	user := partyproto.User{}
	user.Id = userResp["user_id"].(string)
	user.Email = userResp["email"].(string)
	user.Picture = userResp["picture"].(string)
	user.Name = userResp["name"].(string)

	userResponse := partyproto.GetUserResponse{}
	userResponse.User = &user
	return &userResponse, nil
}

// ChangePassword - used to update password
func (u *UserService) ChangePassword(ctx context.Context, in *partyproto.ChangePasswordRequest) (*partyproto.ChangePasswordResponse, error) {
	// https://{yourDomain}/dbconnections/change_password \
	// url := "https://dev-llzybv0gk4ybnuqm.us.auth0.com/dbconnections/change_password"
	url := "https://" + u.ServerOptions.Auth0Domain + "/dbconnections/change_password"

	fmt.Println("url", url)
	// payload := strings.NewReader(`{"client_id":"lJUrHf9acvvqhA9TMaLJJ87XB4pUJTaI","email":"sprov300@gmail.com", "connection":"Username-Password-Authentication"}`)

	payload := strings.NewReader(`{"client_id":"` + u.ServerOptions.Auth0ClientId + `","email":"` + in.Email + `", "connection":"` + u.ServerOptions.Auth0Connection + `"}`)
	fmt.Println("payload", payload)

	req, _ := http.NewRequest("POST", url, payload)

	req.Header.Add("content-type", "application/json")
	req.Header.Add("cache-control", "no-cache")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer res.Body.Close()

	return &partyproto.ChangePasswordResponse{}, nil
}

// DeleteUser - used to get user by Id
func (u *UserService) DeleteUser(ctx context.Context, in *partyproto.DeleteUserRequest) (*partyproto.DeleteUserResponse, error) {
	// url := "https://dev-q3i1a0rds2l1nrpb.us.auth0.com/api/v2/users/" + userID
	url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/users/" + in.UserId
	_, err := common.SendRequest("DELETE", url, nil, "Bearer "+u.ServerOptions.Auth0MgmtToken)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}

	return &partyproto.DeleteUserResponse{}, nil
}

// RefreshToken - used for Refresh Token
// this needs to convert to auth0 code
func (u *UserService) RefreshToken(ctx context.Context, form *partyproto.RefreshTokenRequest) (*partyproto.RefreshTokenResponse, error) {
	return nil, nil
}

func (u *UserService) UpdateUser(ctx context.Context, form *partyproto.UpdateUserRequest) (*partyproto.UpdateUserResponse, error) {
	return nil, nil
}

// GetAuthUserDetails - used to get auth user details
func (u *UserService) GetAuthUserDetails(ctx context.Context, in *partyproto.GetAuthUserDetailsRequest) (*partyproto.GetAuthUserDetailsResponse, error) {
	select {
	case <-ctx.Done():
		err := errors.New("Client closed connection")
		u.log.Error("Error", zap.Int("msgnum", 1583), zap.Error(err))
		return nil, err
	default:
		fmt.Println("UserService GetAuthUserDetails()")
		fmt.Println("UserService GetAuthUserDetails() in.TokenString", in.TokenString)
		// user := partyproto.User{}
		resp, err := u.RedisService.Get(in.TokenString)
		if err != nil {
			u.log.Error("Error", zap.Int("msgnum", 1583), zap.Error(err))
		}
		fmt.Println("UserService GetAuthUserDetails() resp err", err)
		fmt.Println("UserService GetAuthUserDetails() resp", resp)
		v := partyproto.GetAuthUserDetailsResponse{}
		if resp == "" {
			// url := "https://dev-llzybv0gk4ybnuqm.us.auth0.com/userinfo"
			url := "https://" + u.ServerOptions.Auth0Domain + "/userinfo"
			fmt.Println("url", url)
			respBody, err := common.SendRequest("GET", url, nil, "Bearer "+in.TokenString)
			if err != nil {
				fmt.Println("err", err)
				return nil, err
			}
			fmt.Println("UserService GetAuthUserDetails() string(respBody) is", string(respBody))

			jsonDataReader := strings.NewReader(string(respBody))
			decoder := json.NewDecoder(jsonDataReader)
			var userDetail map[string]interface{}
			err = decoder.Decode(&userDetail)
			if err != nil {
				fmt.Println("err", err)
				return nil, err
			}
			fmt.Println("UserService GetAuthUserDetails() userDetail", userDetail)
			fmt.Println("UserService GetAuthUserDetails() userDetail", userDetail["sub"])
			fmt.Println("UserService GetAuthUserDetails() userDetail", userDetail["email"])

			v.Email = userDetail["sub"].(string)
			v.UserId = userDetail["email"].(string)

		} else {

			err = json.Unmarshal([]byte(resp), &v)
			if err != nil {
				u.log.Error("Error", zap.Int("msgnum", 266), zap.Error(err))
			}

		}

		v.RequestId = common.GetRequestID()
		fmt.Println("UserService GetAuthUserDetails() v is", v)
		fmt.Println("UserService GetAuthUserDetails() end")
		return &v, nil
	}
}

// CreateRole - used to create Role
func (u *UserService) CreateRole(ctx context.Context, in *partyproto.CreateRoleRequest) (*partyproto.CreateRoleResponse, error) {
	// url := "https://dev-llzybv0gk4ybnuqm.us.auth0.com/api/v2/roles"
	/*url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/roles"
	payload := strings.NewReader(`{"name":"` + in.Name + `","description":"` + in.Description + `"}`)

	_, err := common.SendRequest("POST", url, payload, "Bearer "+u.ServerOptions.Auth0MgmtToken)*/
	inReq := in.CreateRole
	inReq.Auth0Domain = u.ServerOptions.Auth0Domain
	inReq.Auth0MgmtToken = u.ServerOptions.Auth0MgmtToken
	role, err := common.CreateRoleResp(ctx, inReq)
	if err != nil {
		log.Error("Error", zap.String("user", inReq.UserEmail), zap.String("reqid", inReq.RequestId), zap.Error(err))
		return nil, err
	}
	roleResponse := partyproto.CreateRoleResponse{}
	roleResponse.Role = role

	return &partyproto.CreateRoleResponse{}, nil
}

// GetRole - used to get Role
func (u *UserService) GetRole(ctx context.Context, in *partyproto.GetRoleRequest) (*partyproto.GetRoleResponse, error) {
	// url := "https://dev-llzybv0gk4ybnuqm.us.auth0.com/api/v2/roles/" + roleID
	/*url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/roles/" + in.RoleId
	respBody, err := common.SendRequest("GET", url, nil, "Bearer "+u.ServerOptions.Auth0MgmtToken)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}

	jsonDataReader := strings.NewReader(string(respBody))
	decoder := json.NewDecoder(jsonDataReader)
	var roleResp map[string]interface{}
	err = decoder.Decode(&roleResp)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}
	role := commonproto.Role{}
	role.Id = roleResp["id"].(string)
	role.Name = roleResp["name"].(string)
	role.Description = roleResp["description"].(string)*/
	inReq := in.GetRole
	inReq.Auth0Domain = u.ServerOptions.Auth0Domain
	inReq.Auth0MgmtToken = u.ServerOptions.Auth0MgmtToken
	role, err := common.GetRoleResp(ctx, inReq)
	if err != nil {
		log.Error("Error", zap.String("user", inReq.UserEmail), zap.String("reqid", inReq.RequestId), zap.Error(err))
		return nil, err
	}
	roleResponse := partyproto.GetRoleResponse{}
	roleResponse.Role = role
	return &roleResponse, nil
}

// DeleteRole - used to delete Role
func (u *UserService) DeleteRole(ctx context.Context, in *partyproto.DeleteRoleRequest) (*partyproto.DeleteRoleResponse, error) {
	// url := "https://dev-llzybv0gk4ybnuqm.us.auth0.com/api/v2/roles/" + roleID
	/*url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/roles/" + in.RoleId
	_, err := common.SendRequest("DELETE", url, nil, "Bearer "+u.ServerOptions.Auth0MgmtToken)*/
	inReq := in.DeleteRole
	inReq.Auth0Domain = u.ServerOptions.Auth0Domain
	inReq.Auth0MgmtToken = u.ServerOptions.Auth0MgmtToken
	err := common.DeleteRoleResp(ctx, inReq)
	if err != nil {
		log.Error("Error", zap.String("user", inReq.UserEmail), zap.String("reqid", inReq.RequestId), zap.Error(err))
		return nil, err
	}
	return &partyproto.DeleteRoleResponse{}, nil
}

// UpdateRole - used to update Role
func (u *UserService) UpdateRole(ctx context.Context, in *partyproto.UpdateRoleRequest) (*partyproto.UpdateRoleResponse, error) {
	// url := "https://dev-llzybv0gk4ybnuqm.us.auth0.com/api/v2/roles/"   + roleID
	/*url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/roles/" + in.RoleId
	payload := strings.NewReader(`{"name":"` + in.Name + `","description":"` + in.Description + `"}`)

	_, err := common.SendRequest("PATCH", url, payload, "Bearer "+u.ServerOptions.Auth0MgmtToken)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}*/

	inReq := in.UpdateRole
	inReq.Auth0Domain = u.ServerOptions.Auth0Domain
	inReq.Auth0MgmtToken = u.ServerOptions.Auth0MgmtToken
	role, err := common.UpdateRoleResp(ctx, inReq)
	if err != nil {
		log.Error("Error", zap.String("user", inReq.UserEmail), zap.String("reqid", inReq.RequestId), zap.Error(err))
		return nil, err
	}
	roleResponse := partyproto.UpdateRoleResponse{}
	roleResponse.Role = role

	return &roleResponse, nil
}

// GetRoles - used to get Roles
func (u *UserService) GetRoles(ctx context.Context, in *partyproto.GetRolesRequest) (*partyproto.GetRolesResponse, error) {
	// url := "https://dev-llzybv0gk4ybnuqm.us.auth0.com/api/v2/roles
	/*url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/roles"
	respBody, err := common.SendRequest("GET", url, nil, "Bearer "+u.ServerOptions.Auth0MgmtToken)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}
	jsonDataReader := strings.NewReader(string(respBody))
	decoder := json.NewDecoder(jsonDataReader)
	var roleResp []map[string]interface{}
	err = decoder.Decode(&roleResp)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}
	roles := []*commonproto.Role{}
	for _, rl := range roleResp {
		role := commonproto.Role{}
		role.Id = rl["id"].(string)
		role.Name = rl["name"].(string)
		role.Description = rl["description"].(string)
		roles = append(roles, &role)
	}*/
	inReq := in.GetRoles
	inReq.Auth0Domain = u.ServerOptions.Auth0Domain
	inReq.Auth0MgmtToken = u.ServerOptions.Auth0MgmtToken
	roles, err := common.GetRolesResp(ctx, inReq)
	if err != nil {
		log.Error("Error", zap.String("user", inReq.UserEmail), zap.String("reqid", inReq.RequestId), zap.Error(err))
		return nil, err
	}
	roleResponse := partyproto.GetRolesResponse{}
	roleResponse.Roles = roles
	return &roleResponse, nil
}

func (u *UserService) AddPermisionsToRoles(ctx context.Context, in *partyproto.AddPermisionsToRolesRequest) (*partyproto.AddPermisionsToRolesResponse, error) {
	// url := "https://dev-llzybv0gk4ybnuqm.us.auth0.com/api/v2/roles/" + roleID + "/permissions"
	/*url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/roles/" + in.RoleId + "/permissions"

	payload := strings.NewReader(`{"permissions": [{"resource_server_identifier":"` + in.ResourceServerIdentifier + `","permission_name":"` + in.PermissionName + `"}]}`)

	_, err := common.SendRequest("POST", url, payload, "Bearer "+u.ServerOptions.Auth0MgmtToken)*/

	inReq := in.AddPermisionsToRoles
	inReq.Auth0Domain = u.ServerOptions.Auth0Domain
	inReq.Auth0MgmtToken = u.ServerOptions.Auth0MgmtToken
	err := common.AddPermisionsToRolesResp(ctx, inReq)
	if err != nil {
		log.Error("Error", zap.String("user", inReq.UserEmail), zap.String("reqid", inReq.RequestId), zap.Error(err))
		return nil, err
	}

	return &partyproto.AddPermisionsToRolesResponse{}, nil
}

func (u *UserService) RemoveRolePermission(ctx context.Context, in *partyproto.RemoveRolePermissionRequest) (*partyproto.RemoveRolePermissionResponse, error) {
	// url := "https://dev-llzybv0gk4ybnuqm.us.auth0.com/api/v2/roles/" + roleID + "/permissions"
	/*url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/roles/" + in.RoleId + "/permissions"

	payload := strings.NewReader(`{"permissions": [{"resource_server_identifier":"` + in.ResourceServerIdentifier + `","permission_name":"` + in.PermissionName + `"}]}`)

	_, err := common.SendRequest("DELETE", url, payload, "Bearer "+u.ServerOptions.Auth0MgmtToken)*/

	inReq := in.RemoveRolePermission
	inReq.Auth0Domain = u.ServerOptions.Auth0Domain
	inReq.Auth0MgmtToken = u.ServerOptions.Auth0MgmtToken
	err := common.RemoveRolePermissionResp(ctx, inReq)
	if err != nil {
		log.Error("Error", zap.String("user", inReq.UserEmail), zap.String("reqid", inReq.RequestId), zap.Error(err))
		return nil, err
	}

	return &partyproto.RemoveRolePermissionResponse{}, nil
}

func (u *UserService) GetRolePermissions(ctx context.Context, in *partyproto.GetRolePermissionsRequest) (*partyproto.GetRolePermissionsResponse, error) {
	// url := "https://dev-llzybv0gk4ybnuqm.us.auth0.com/api/v2/roles/" + roleID + "/permissions"
	/*url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/roles/" + in.RoleId + "/permissions"

	respBody, err := common.SendRequest("GET", url, nil, "Bearer "+u.ServerOptions.Auth0MgmtToken)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}

	jsonDataReader := strings.NewReader(string(respBody))
	decoder := json.NewDecoder(jsonDataReader)
	var rolePermissionResp []map[string]interface{}
	err = decoder.Decode(&rolePermissionResp)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}
	rolePermissions := []*partyproto.RolePermission{}
	for _, rl := range rolePermissionResp {
		rolePermission := partyproto.RolePermission{}
		rolePermission.PermissionName = rl["permission_name"].(string)
		rolePermission.Description = rl["description"].(string)
		rolePermission.ResourceServerName = rl["resource_server_name"].(string)
		rolePermission.ResourceServerIdentifier = rl["resource_server_identifier"].(string)
		rolePermissions = append(rolePermissions, &rolePermission)
	}*/

	inReq := in.GetRolePermissions
	inReq.Auth0Domain = u.ServerOptions.Auth0Domain
	inReq.Auth0MgmtToken = u.ServerOptions.Auth0MgmtToken
	rolePermissions, err := common.GetRolePermissionsResp(ctx, inReq)
	if err != nil {
		log.Error("Error", zap.String("user", inReq.UserEmail), zap.String("reqid", inReq.RequestId), zap.Error(err))
		return nil, err
	}

	rolePermissionsResponse := partyproto.GetRolePermissionsResponse{}
	rolePermissionsResponse.RolePermissions = rolePermissions
	return &rolePermissionsResponse, nil
}

// AssignRolesToUsers - used to assign roles to users
func (u *UserService) AssignRolesToUsers(ctx context.Context, in *partyproto.AssignRolesToUsersRequest) (*partyproto.AssignRolesToUsersResponse, error) {
	// url := "https://dev-llzybv0gk4ybnuqm.us.auth0.com/api/v2/users/" + userID + "/roles"

	/*url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/users/" + in.AssignToUserId + "/roles"

	payload := strings.NewReader(`{ "roles": [ "` + in.RoleId + `"] }`)

	_, err := common.SendRequest("POST", url, payload, "Bearer "+u.ServerOptions.Auth0MgmtToken)*/
	inReq := in.AssignRolesToUsers
	inReq.Auth0Domain = u.ServerOptions.Auth0Domain
	inReq.Auth0MgmtToken = u.ServerOptions.Auth0MgmtToken
	err := common.AssignRolesToUsersResp(ctx, inReq)
	if err != nil {
		log.Error("Error", zap.String("user", inReq.UserEmail), zap.String("reqid", inReq.RequestId), zap.Error(err))
		return nil, err
	}
	return &partyproto.AssignRolesToUsersResponse{}, nil
}

// ViewUserRoles - used to View User Roles
func (u *UserService) ViewUserRoles(ctx context.Context, in *partyproto.ViewUserRolesRequest) (*partyproto.ViewUserRolesResponse, error) {
	/*url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/users/" + in.UserId + "/roles"
	respBody, err := common.SendRequest("GET", url, nil, "Bearer "+u.ServerOptions.Auth0MgmtToken)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}

	jsonDataReader := strings.NewReader(string(respBody))
	decoder := json.NewDecoder(jsonDataReader)
	var roleResp []map[string]interface{}
	err = decoder.Decode(&roleResp)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}
	roles := []*commonproto.Role{}
	for _, rl := range roleResp {
		role := commonproto.Role{}
		role.Id = rl["id"].(string)
		role.Name = rl["name"].(string)
		role.Description = rl["description"].(string)
		roles = append(roles, &role)
	}*/

	inReq := in.ViewUserRoles
	inReq.Auth0Domain = u.ServerOptions.Auth0Domain
	inReq.Auth0MgmtToken = u.ServerOptions.Auth0MgmtToken
	roles, err := common.ViewUserRolesResp(ctx, inReq)
	if err != nil {
		log.Error("Error", zap.String("user", inReq.UserEmail), zap.String("reqid", inReq.RequestId), zap.Error(err))
		return nil, err
	}
	roleResponse := partyproto.ViewUserRolesResponse{}
	roleResponse.Roles = roles
	return &roleResponse, nil
}

func (u *UserService) GetRoleUsers(ctx context.Context, in *partyproto.GetRoleUsersRequest) (*partyproto.GetRoleUsersResponse, error) {
	url := "https://" + u.ServerOptions.Auth0Domain + "/api/v2/roles/" + in.RoleId + "/users"

	respBody, err := common.SendRequest("GET", url, nil, "Bearer "+u.ServerOptions.Auth0MgmtToken)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}

	jsonDataReader := strings.NewReader(string(respBody))
	decoder := json.NewDecoder(jsonDataReader)
	var roleUsers []map[string]interface{}
	err = decoder.Decode(&roleUsers)
	if err != nil {
		fmt.Println("err", err)
		return nil, err
	}
	users := []*partyproto.User{}
	for _, usr := range roleUsers {
		user := partyproto.User{}
		user.Id = usr["user_id"].(string)
		user.Email = usr["email"].(string)
		user.Picture = usr["picture"].(string)
		user.Name = usr["name"].(string)
		users = append(users, &user)
	}
	roleUsersResponse := partyproto.GetRoleUsersResponse{}
	roleUsersResponse.Users = users
	return &roleUsersResponse, nil
}

// GetUserWithNewContext - GetUserWithNewContext
func GetUserWithNewContext(ctx context.Context, userId string, userEmail string, requestId string, userServiceClient partyproto.UserServiceClient) (*partyproto.User, error) {
	getRequest := commonproto.GetRequest{}
	getRequest.Id = userId
	getRequest.UserEmail = userEmail
	getRequest.RequestId = requestId
	fmt.Println("partyservices/user_service.go:GetUserWithNewContext() getRequest is", getRequest)
	ctxNew, err := common.CreateCtxJWT(ctx)
	if err != nil {
		log.Error("Error", zap.String("user", userEmail), zap.String("reqid", requestId), zap.Int("msgnum", 4319), zap.Error(err))
		return nil, err
	}
	fmt.Println("partyservices/user_service.go:GetUserWithNewContext() ctxNew is", ctxNew)
	fmt.Println("partyservices/user_service.go:GetUserWithNewContext() userServiceClient is", userServiceClient)
	form := partyproto.GetUserRequest{}
	form.GetRequest = &getRequest
	userResponse, err := userServiceClient.GetUser(ctxNew, &form)
	fmt.Println("partyservices/user_service.go:GetUserWithNewContext() userResponse is", userResponse)
	fmt.Println("partyservices/user_service.go:GetUserWithNewContext() err is", err)
	if err != nil {
		log.Error("Error", zap.String("user", userEmail), zap.String("reqid", requestId), zap.Int("msgnum", 4319), zap.Error(err))
		return nil, err
	}
	user := userResponse.User
	fmt.Println("partyservices/user_service.go:GetUserWithNewContext() user is", user)
	return user, nil
}

func (u *UserService) UserTracer(ctx context.Context) {
	/*_, span := tracer.Start(ctx, "workHard",
		trace.WithAttributes(attribute.String("extra.key", "extra.value")))
	defer span.End()

	time.Sleep(50 * time.Millisecond)*/
	dsn := "https://iTlkx26v8LV0NMwB1cL8Dw@api.uptrace.dev?grpc=4317"
	if dsn == "" {
		panic("UPTRACE_DSN environment variable is required")
	}
	fmt.Println("using DSN:", dsn)

	creds := credentials.NewClientTLSFromCert(nil, "")
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint("otlp.uptrace.dev:4317"),
		otlptracegrpc.WithTLSCredentials(creds),
		otlptracegrpc.WithHeaders(map[string]string{
			// Set the Uptrace DSN here or use UPTRACE_DSN env var.
			"uptrace-dsn": dsn,
		}),
		otlptracegrpc.WithCompressor(gzip.Name),
	)
	if err != nil {
		panic(err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter,
		sdktrace.WithMaxQueueSize(10_000),
		sdktrace.WithMaxExportBatchSize(10_000))
	// Call shutdown to flush the buffers when program exits.
	defer bsp.Shutdown(ctx)

	resource, err := resource.New(ctx,
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			attribute.String("service.name", "myservice"),
			attribute.String("service.version", "1.0.0"),
		))
	/*resource, err := resource.New(ctx,
	resource.WithFromEnv(),
	resource.WithTelemetrySDK(),
	resource.WithHost(),
	resource.WithAttributes(
		attribute.String("service.name", "userservice"),
		attribute.String("service.version", "1.0.0"),
	))*/if err != nil {
		panic(err)
	}

	tracerProvider := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource),
		sdktrace.WithIDGenerator(xray.NewIDGenerator()),
	)
	tracerProvider.RegisterSpanProcessor(bsp)

	// Install our tracer provider and we are done.
	otel.SetTracerProvider(tracerProvider)

	tracer := otel.Tracer("myservice")
	// tracer := otel.Tracer("userservice")
	ctx, span := tracer.Start(ctx, "UserTracer",
		trace.WithAttributes(attribute.String("extra.key", "extra.value")))
	defer span.End()
}
