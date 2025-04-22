package web

import (
	svc "github.com/panyam/leetcoach/services"
	oa "github.com/panyam/oneauth"
	"golang.org/x/oauth2"
)

type LCAuthUser struct {
	*svc.User
}

func (n *LCAuthUser) Id() string {
	return n.User.Id
}

func (n *LCAuthUser) Profile() map[string]any {
	return n.User.Profile.Properties
}

func (n *LCApp) GetUserByID(userId string) (oa.User, error) {
	var user LCAuthUser
	var err error
	if userId == "test1" {
		// Mocking user login
		return &LCAuthUser{
			User: &svc.User{
				Id: "test1",
				Profile: svc.StringMapField{
					Properties: map[string]any{
						"Name": "Test User",
					},
				},
			},
		}, nil
	}
	user.User, err = n.ClientMgr.GetAuthService().GetUserByID(userId)
	return &user, err
}

func (n *LCApp) EnsureAuthUser(authtype string, provider string, token *oauth2.Token, userInfo map[string]any) (oa.User, error) {
	var user LCAuthUser
	var err error
	// Mocking user login
	email := userInfo["email"].(string)
	if email == "test@gmail.com" {
		return &LCAuthUser{
			User: &svc.User{
				Id: "test1",
				Profile: svc.StringMapField{
					Properties: map[string]any{
						"Name": "Test User",
					},
				},
			},
		}, nil
	}
	user.User, err = n.ClientMgr.GetAuthService().EnsureAuthUser(authtype, provider, token, userInfo)
	return &user, err
}

func (n *LCApp) ValidateUsernamePassword(username string, password string) (out oa.User, err error) {
	if username == "test@gmail.com" {
		out = &LCAuthUser{
			User: &svc.User{
				Id: "test1",
				Profile: svc.StringMapField{
					Properties: map[string]any{
						"Name": "Test User",
					},
				},
			},
		}
	}
	return
}
