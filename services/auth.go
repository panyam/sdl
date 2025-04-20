package services

import (
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/oauth2"
)

type AuthService struct {
	clients *ClientMgr
}

func (a *AuthService) GetUserByID(userId string) (*User, error) {
	user := User{}
	err := a.clients.GetUserDSClient().GetByID(userId, &user)
	return &user, err
}

func (a *AuthService) EnsureAuthUser(authtype string, provider string, token *oauth2.Token, userInfo map[string]any) (user *User, err error) {
	slog.Info("EnsuringUser: ", "user", userInfo)
	// fullName := fmt.Sprintf("%s %s", userInfo["given_name"].(string), userInfo["family_name"].(string))
	// userPicture := userInfo["picture"].(string)
	userEmail := userInfo["email"].(string)

	idsc := a.clients.GetIdentityDSClient()
	idKey := fmt.Sprintf("email:%s", userEmail)
	identity := Identity{
		IsActive: true,
		BaseModel: BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}
	err = idsc.GetByID(idKey, &identity)
	if err != nil {
		identity.IdentityType = "email"
		identity.IdentityKey = userEmail
		if _, err = idsc.SaveEntity(&identity); err != nil {
			slog.Error("Error saving identity: ", "err", err)
			return
		}
	}

	userdsc := a.clients.GetUserDSClient()
	if identity.HasUser() {
		user = &User{}
		err := userdsc.GetByID(identity.PrimaryUser, user)
		if err != nil {
			user = nil
			slog.Error("Error getting user: ", "user", identity.PrimaryUser, "err", err)
		} else {
			/*
				} else if user.Picture != userPicture || user.Name != fullName || user.Email != idKey {
					// chance to update
					user.Picture = userPicture
					user.Name = fullName
					user.Email = userEmail
					user.Id = idKey
					if _, err = userdsc.SaveEntity(&user); err != nil {
						slog.Error("Error saving user: ", idKey, err)
						ctx.JSON(http.StatusUnauthorized, map[string]any{"error": "Unable to save user"})
						return
					}
			*/
		}
	}

	// was was not found so create one
	if user == nil {
		user = &User{
			Profile:  StringMapField{Properties: userInfo},
			IsActive: true,
			BaseModel: BaseModel{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
		}
		// create one
		if _, err = userdsc.SaveEntity(user); err != nil {
			slog.Error("Error saving user: ", "user", user, "err", err)
			return
		}
		slog.Info("Saved User: ", "user", user)
		identity.PrimaryUser = user.Id
		if _, err = idsc.SaveEntity(&identity); err != nil {
			slog.Error("Error saving identity: ", "error", err)
			return
		}
	}

	channeldsc := a.clients.GetChannelDSClient()
	channel := Channel{
		Provider:    provider,
		IdentityKey: identity.Key(),
		BaseModel: BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		Credentials: StringMapField{Properties: map[string]any{
			"access_token":  token.AccessToken,
			"refresh_token": token.RefreshToken,
			"token_type":    token.TokenType,
			"expiry":        token.Expiry,
		}},
		Profile: StringMapField{Properties: userInfo},
	}
	if channeldsc.GetByID(channel.Key(), &channel) != nil {
		// then create it
		if _, err = channeldsc.SaveEntity(&channel); err != nil {
			slog.Error("Error saving identity: ", "error", err)
			return
		}
	}

	if !channel.HasIdentity() {
		channel.IdentityKey = identity.Key()
		if _, err = channeldsc.SaveEntity(&channel); err != nil {
			slog.Error("Error saving identity: ", "error", err)
			return
		}
	}

	// Now validate the channel
	return
}
