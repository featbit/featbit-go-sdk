package interfaces

import (
	"fmt"
	"strings"
)

const (
	key      = "key"
	userName = "name"
)

var builtins = map[string]string{
	"key":   key,
	"keyid": key,
	"name":  userName,
}

type FBUser struct {
	userName string
	key      string
	custom   map[string]string
}

func (u *FBUser) IsValid() bool {
	if u.key == "" || u.userName == "" {
		return false
	}
	return true
}

func (u *FBUser) GetKey() string {
	return u.key
}

func (u *FBUser) GetUserName() string {
	return u.userName
}

func (u *FBUser) CustomAttributes() map[string]string {
	attrs := make(map[string]string, len(u.custom))
	for k, v := range u.custom {
		attrs[k] = v
	}
	return attrs
}

func (u *FBUser) Get(attribute string) string {
	attr := strings.ToLower(attribute)
	switch builtins[attr] {
	case key:
		return u.key
	case userName:
		return u.userName
	default:
		return u.custom[attribute]
	}
}

type UserBuilder interface {
	Key(value string) UserBuilder
	UserName(value string) UserBuilder
	Custom(attribute string, value string) UserBuilder
	Build() (FBUser, error)
}

type userBuilderImpl struct {
	userName string
	key      string
	custom   map[string]string
}

func NewUserBuilder(key string) UserBuilder {
	return &userBuilderImpl{key: key, userName: key, custom: make(map[string]string)}
}

func (u *userBuilderImpl) Key(value string) UserBuilder {
	u.key = value
	return u
}

func (u *userBuilderImpl) UserName(value string) UserBuilder {
	u.userName = value
	return u
}

func (u *userBuilderImpl) Custom(attribute string, value string) UserBuilder {
	u.custom[attribute] = value
	return u
}

func (u *userBuilderImpl) Build() (FBUser, error) {
	if u.key == "" {
		return FBUser{}, fmt.Errorf("key shouldn't be empty")
	}

	if u.userName == "" {
		return FBUser{}, fmt.Errorf("user name shouldn't be empty")
	}

	user := FBUser{
		key:      u.key,
		userName: u.userName,
		custom:   make(map[string]string, len(u.custom)),
	}
	for k, v := range u.custom {
		user.custom[k] = v
	}
	return user, nil
}
