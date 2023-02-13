package featbit

import (
	"github.com/featbit/featbit-go-sdk/interfaces"
	"github.com/featbit/featbit-go-sdk/internal/types/insight"
)

func convertFBUserToEventUser(user *interfaces.FBUser) insight.EventUser {
	ret := insight.EventUser{
		KeyId: user.GetKey(),
		Name:  user.GetUserName(),
		Attrs: make([]insight.UserAttribute, len(user.CustomAttributes())),
	}
	for k, v := range user.CustomAttributes() {
		ret.Attrs = append(ret.Attrs, insight.UserAttribute{Name: k, Value: v})
	}
	return ret
}
