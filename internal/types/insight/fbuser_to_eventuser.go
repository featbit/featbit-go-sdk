package insight

import (
	. "github.com/featbit/featbit-go-sdk/interfaces"
)

func ConvertFBUserToEventUser(user *FBUser) EventUser {
	ret := EventUser{
		KeyId: user.GetKey(),
		Name:  user.GetUserName(),
		Attrs: make([]UserAttribute, len(user.CustomAttributes())),
	}
	for k, v := range user.CustomAttributes() {
		ret.Attrs = append(ret.Attrs, UserAttribute{Name: k, Value: v})
	}
	return ret
}
