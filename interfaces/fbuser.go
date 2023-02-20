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

// FBUser is a collection of attributes that can affect flag evaluation, usually corresponding to a user of your application.
//
// The mandatory properties are the key and name.
//
// The key must uniquely identify each user in an environment;
// this could be a username or email address for authenticated users, or an ID for anonymous users.
//
// The name is used to search your user quickly in feature flag center.
//
// The custom properties are optional, you may also define custom properties with arbitrary names and values.
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

// GetKey returns user's unique key
func (u *FBUser) GetKey() string {
	return u.key
}

// GetUserName returns user's name
func (u *FBUser) GetUserName() string {
	return u.userName
}

// CustomAttributes Returns a copy of all custom attributes set for this user
func (u *FBUser) CustomAttributes() map[string]string {
	attrs := make(map[string]string, len(u.custom))
	for k, v := range u.custom {
		attrs[k] = v
	}
	return attrs
}

// Get gets the value of a user attribute, if present.
//
// This can be either a built-in attribute(key/userName) or a custom one
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

// NewUserBuilder that helps construct FBUser.
//
// The calls can be chained, supporting the following pattern:
// 		user, _ := NewUserBuilder("key").UserName("name").Custom("property", "value").Build()
func NewUserBuilder(key string) UserBuilder {
	return &userBuilderImpl{key: key, userName: key, custom: make(map[string]string)}
}

// Key sets the user's key.
func (u *userBuilderImpl) Key(value string) UserBuilder {
	u.key = value
	return u
}

// UserName sets the user's userName.
func (u *userBuilderImpl) UserName(value string) UserBuilder {
	u.userName = value
	return u
}

// Custom adds a String-valued custom attribute. When set to one of the built-in user attribute keys,
// the key/value pair will be ignored.
func (u *userBuilderImpl) Custom(attribute string, value string) UserBuilder {
	u.custom[attribute] = value
	return u
}

// Build builds the configured FBUser object.
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
