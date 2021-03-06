package control

import (
	"github.com/Myriad-Dreamin/go-model-traits/gorm-crud-dao"
	"github.com/Myriad-Dreamin/minimum-lib/controller"
	"github.com/Myriad-Dreamin/minimum-template/model/db-layer"
	"time"
)

type UserService interface {
	UserServiceSignatureXXX() interface{}
	ListUsers(c controller.MContext)
	Login(c controller.MContext)
	Register(c controller.MContext)
	ChangePassword(c controller.MContext)
	InspectUser(c controller.MContext)
	GetUser(c controller.MContext)
	PutUser(c controller.MContext)
	Delete(c controller.MContext)
}
type ListUsersRequest = gorm_crud_dao.Filter

type ListUsersReply struct {
	Code  int             `json:"code" form:"code"`
	Users []ListUserReply `json:"users" form:"users"`
}

type ListUserReply struct {
	NickName  string    `json:"nick_name" form:"nick_name"`
	LastLogin time.Time `form:"last_login" json:"last_login"`
}

type LoginRequest struct {
	Id       uint   `json:"id" form:"id"`
	NickName string `json:"nick_name" form:"nick_name"`
	Phone    string `json:"phone" form:"phone"`
	Password string `form:"password" binding:"required" json:"password"`
}

type LoginReply struct {
	Code         int      `json:"code" form:"code"`
	Id           uint     `json:"id" form:"id"`
	Phone        string   `json:"phone" form:"phone"`
	NickName     string   `json:"nick_name" form:"nick_name"`
	Name         string   `json:"name" form:"name"`
	Identity     []string `json:"identity" form:"identity"`
	Token        string   `json:"token" form:"token"`
	RefreshToken string   `json:"refresh_token" form:"refresh_token"`
}

type RegisterRequest struct {
	Name     string `json:"name" form:"name" binding:"required"`
	NickName string `binding:"required" json:"nick_name" form:"nick_name"`
	Phone    string `binding:"required" json:"phone" form:"phone"`
	Password string `json:"password" form:"password" binding:"required"`
}

type RegisterReply struct {
	Code int  `json:"code" form:"code"`
	Id   uint `json:"id" form:"id"`
}

type ChangePasswordRequest struct {
	OldPassword string `binding:"required" json:"old_password" form:"old_password"`
	NewPassword string `json:"new_password" form:"new_password" binding:"required"`
}

type InspectUserReply struct {
	Code int           `json:"code" form:"code"`
	User *dblayer.User `json:"user" form:"user"`
}

type GetUserReply struct {
	Code      int       `json:"code" form:"code"`
	NickName  string    `json:"nick_name" form:"nick_name"`
	LastLogin time.Time `json:"last_login" form:"last_login"`
}

type PutUserRequest struct {
	Phone string `json:"phone" form:"phone"`
}

func PSerializeListUsersReply(_code int, _users []ListUserReply) *ListUsersReply {

	return &ListUsersReply{
		Code:  _code,
		Users: _users,
	}
}
func SerializeListUsersReply(_code int, _users []ListUserReply) ListUsersReply {

	return ListUsersReply{
		Code:  _code,
		Users: _users,
	}
}
func _packSerializeListUsersReply(_code int, _users []ListUserReply) ListUsersReply {

	return ListUsersReply{
		Code:  _code,
		Users: _users,
	}
}
func PackSerializeListUsersReply(_code []int, _users [][]ListUserReply) (pack []ListUsersReply) {
	for i := range _code {
		pack = append(pack, _packSerializeListUsersReply(_code[i], _users[i]))
	}
	return
}
func PSerializeListUserReply(vUser dblayer.User) *ListUserReply {

	return &ListUserReply{
		NickName:  vUser.NickName,
		LastLogin: vUser.LastLogin,
	}
}
func SerializeListUserReply(vUser dblayer.User) ListUserReply {

	return ListUserReply{
		NickName:  vUser.NickName,
		LastLogin: vUser.LastLogin,
	}
}
func _packSerializeListUserReply(vUser dblayer.User) ListUserReply {

	return ListUserReply{
		NickName:  vUser.NickName,
		LastLogin: vUser.LastLogin,
	}
}
func PackSerializeListUserReply(vUser []dblayer.User) (pack []ListUserReply) {
	for i := range vUser {
		pack = append(pack, _packSerializeListUserReply(vUser[i]))
	}
	return
}
func PSerializeLoginRequest(user *dblayer.User, _password string) *LoginRequest {

	return &LoginRequest{
		Id:       user.ID,
		NickName: user.NickName,
		Phone:    user.Phone,
		Password: _password,
	}
}
func SerializeLoginRequest(user *dblayer.User, _password string) LoginRequest {

	return LoginRequest{
		Id:       user.ID,
		NickName: user.NickName,
		Phone:    user.Phone,
		Password: _password,
	}
}
func _packSerializeLoginRequest(user *dblayer.User, _password string) LoginRequest {

	return LoginRequest{
		Id:       user.ID,
		NickName: user.NickName,
		Phone:    user.Phone,
		Password: _password,
	}
}
func PackSerializeLoginRequest(user []*dblayer.User, _password []string) (pack []LoginRequest) {
	for i := range user {
		pack = append(pack, _packSerializeLoginRequest(user[i], _password[i]))
	}
	return
}
func PSerializeLoginReply(_code int, user *dblayer.User, _identity []string, _token string, _refreshToken string) *LoginReply {

	return &LoginReply{
		Code:         _code,
		Id:           user.ID,
		Phone:        user.Phone,
		NickName:     user.NickName,
		Name:         user.Name,
		Identity:     _identity,
		Token:        _token,
		RefreshToken: _refreshToken,
	}
}
func SerializeLoginReply(_code int, user *dblayer.User, _identity []string, _token string, _refreshToken string) LoginReply {

	return LoginReply{
		Code:         _code,
		Id:           user.ID,
		Phone:        user.Phone,
		NickName:     user.NickName,
		Name:         user.Name,
		Identity:     _identity,
		Token:        _token,
		RefreshToken: _refreshToken,
	}
}
func _packSerializeLoginReply(_code int, user *dblayer.User, _identity []string, _token string, _refreshToken string) LoginReply {

	return LoginReply{
		Code:         _code,
		Id:           user.ID,
		Phone:        user.Phone,
		NickName:     user.NickName,
		Name:         user.Name,
		Identity:     _identity,
		Token:        _token,
		RefreshToken: _refreshToken,
	}
}
func PackSerializeLoginReply(_code []int, user []*dblayer.User, _identity [][]string, _token []string, _refreshToken []string) (pack []LoginReply) {
	for i := range _code {
		pack = append(pack, _packSerializeLoginReply(_code[i], user[i], _identity[i], _token[i], _refreshToken[i]))
	}
	return
}
func PSerializeRegisterRequest(user *dblayer.User, _password string) *RegisterRequest {

	return &RegisterRequest{
		Name:     user.Name,
		NickName: user.NickName,
		Phone:    user.Phone,
		Password: _password,
	}
}
func SerializeRegisterRequest(user *dblayer.User, _password string) RegisterRequest {

	return RegisterRequest{
		Name:     user.Name,
		NickName: user.NickName,
		Phone:    user.Phone,
		Password: _password,
	}
}
func _packSerializeRegisterRequest(user *dblayer.User, _password string) RegisterRequest {

	return RegisterRequest{
		Name:     user.Name,
		NickName: user.NickName,
		Phone:    user.Phone,
		Password: _password,
	}
}
func PackSerializeRegisterRequest(user []*dblayer.User, _password []string) (pack []RegisterRequest) {
	for i := range user {
		pack = append(pack, _packSerializeRegisterRequest(user[i], _password[i]))
	}
	return
}
func PSerializeRegisterReply(_code int, user *dblayer.User) *RegisterReply {

	return &RegisterReply{
		Code: _code,
		Id:   user.ID,
	}
}
func SerializeRegisterReply(_code int, user *dblayer.User) RegisterReply {

	return RegisterReply{
		Code: _code,
		Id:   user.ID,
	}
}
func _packSerializeRegisterReply(_code int, user *dblayer.User) RegisterReply {

	return RegisterReply{
		Code: _code,
		Id:   user.ID,
	}
}
func PackSerializeRegisterReply(_code []int, user []*dblayer.User) (pack []RegisterReply) {
	for i := range _code {
		pack = append(pack, _packSerializeRegisterReply(_code[i], user[i]))
	}
	return
}
func PSerializeChangePasswordRequest(_oldPassword string, _newPassword string) *ChangePasswordRequest {

	return &ChangePasswordRequest{
		OldPassword: _oldPassword,
		NewPassword: _newPassword,
	}
}
func SerializeChangePasswordRequest(_oldPassword string, _newPassword string) ChangePasswordRequest {

	return ChangePasswordRequest{
		OldPassword: _oldPassword,
		NewPassword: _newPassword,
	}
}
func _packSerializeChangePasswordRequest(_oldPassword string, _newPassword string) ChangePasswordRequest {

	return ChangePasswordRequest{
		OldPassword: _oldPassword,
		NewPassword: _newPassword,
	}
}
func PackSerializeChangePasswordRequest(_oldPassword []string, _newPassword []string) (pack []ChangePasswordRequest) {
	for i := range _oldPassword {
		pack = append(pack, _packSerializeChangePasswordRequest(_oldPassword[i], _newPassword[i]))
	}
	return
}
func PSerializeInspectUserReply(_code int, _user *dblayer.User) *InspectUserReply {

	return &InspectUserReply{
		Code: _code,
		User: _user,
	}
}
func SerializeInspectUserReply(_code int, _user *dblayer.User) InspectUserReply {

	return InspectUserReply{
		Code: _code,
		User: _user,
	}
}
func _packSerializeInspectUserReply(_code int, _user *dblayer.User) InspectUserReply {

	return InspectUserReply{
		Code: _code,
		User: _user,
	}
}
func PackSerializeInspectUserReply(_code []int, _user []*dblayer.User) (pack []InspectUserReply) {
	for i := range _code {
		pack = append(pack, _packSerializeInspectUserReply(_code[i], _user[i]))
	}
	return
}
func PSerializeGetUserReply(_code int, user *dblayer.User) *GetUserReply {

	return &GetUserReply{
		Code:      _code,
		NickName:  user.NickName,
		LastLogin: user.LastLogin,
	}
}
func SerializeGetUserReply(_code int, user *dblayer.User) GetUserReply {

	return GetUserReply{
		Code:      _code,
		NickName:  user.NickName,
		LastLogin: user.LastLogin,
	}
}
func _packSerializeGetUserReply(_code int, user *dblayer.User) GetUserReply {

	return GetUserReply{
		Code:      _code,
		NickName:  user.NickName,
		LastLogin: user.LastLogin,
	}
}
func PackSerializeGetUserReply(_code []int, user []*dblayer.User) (pack []GetUserReply) {
	for i := range _code {
		pack = append(pack, _packSerializeGetUserReply(_code[i], user[i]))
	}
	return
}
func PSerializePutUserRequest(user *dblayer.User) *PutUserRequest {

	return &PutUserRequest{
		Phone: user.Phone,
	}
}
func SerializePutUserRequest(user *dblayer.User) PutUserRequest {

	return PutUserRequest{
		Phone: user.Phone,
	}
}
func _packSerializePutUserRequest(user *dblayer.User) PutUserRequest {

	return PutUserRequest{
		Phone: user.Phone,
	}
}
func PackSerializePutUserRequest(user []*dblayer.User) (pack []PutUserRequest) {
	for i := range user {
		pack = append(pack, _packSerializePutUserRequest(user[i]))
	}
	return
}
