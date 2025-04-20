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

func (n *LCApp) GetUserByID(userId string) (oa.User, error) {
	var user LCAuthUser
	var err error
	user.User, err = n.ClientMgr.GetAuthService().GetUserByID(userId)
	return &user, err
}

func (n *LCApp) EnsureAuthUser(authtype string, provider string, token *oauth2.Token, userInfo map[string]any) (oa.User, error) {
	var user LCAuthUser
	var err error
	user.User, err = n.ClientMgr.GetAuthService().EnsureAuthUser(authtype, provider, token, userInfo)
	return &user, err
}
