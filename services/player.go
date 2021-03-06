package services

import (
	"encoding/json"

	"github.com/yasshi2525/RushHour/auth"
	"github.com/yasshi2525/RushHour/entities"
)

// CreatePlayer creates player.
func CreatePlayer(loginid string, displayname string, password string, hue int, lv entities.PlayerType) (*entities.Player, error) {
	o, err := Model.PasswordSignUp(loginid, password, lv)
	if err != nil {
		return nil, err
	}
	o.CustomDisplayName = auther.Encrypt(displayname)
	o.UseCustomDisplayName = true
	o.Hue = hue
	o.UseCustomImage = true
	AddOpLog("CreatePlayer", o)
	return o, nil

}

// OAuthSignIn find or create Player by OAuth
func OAuthSignIn(authType entities.AuthType, info *auth.OAuthInfo) (*entities.Player, error) {
	if o, err := Model.OAuthSignIn(authType, info); err != nil {
		return nil, err
	} else {
		return o, nil
	}
}

// SignOut delete Player's token value
func SignOut(o *entities.Player) {
	o.SignOut()
}

// PasswordSignIn finds Player by loginid and password
func PasswordSignIn(loginid string, password string) (*entities.Player, error) {
	if o, err := Model.PasswordSignIn(loginid, password); err != nil {
		return nil, err
	} else {
		return o, nil
	}
}

// PasswordSignUp creates Player with loginid and password
func PasswordSignUp(loginid string, name string, password string, hue int, lv entities.PlayerType) (*entities.Player, error) {
	o, err := Model.PasswordSignUp(loginid, password, lv)
	if err != nil {
		return nil, err
	}
	o.CustomDisplayName = auther.Encrypt(name)
	o.UseCustomDisplayName = true
	o.Hue = hue
	o.UseCustomImage = true
	return o, nil

}

// AccountSettings returns user customizable attributes.
type AccountSettings struct {
	Player         *entities.Player  `json:"-"`
	CustomName     string            `json:"custom_name"`
	CustomImage    string            `json:"custom_image"`
	AuthType       entities.AuthType `json:"auth_type"`
	LoginID        string            `json:"email,omitempty"`
	OAuthName      string            `json:"oauth_name,omitempty"`
	UseCustomName  bool              `json:"use_cname,omitempty"`
	OAuthImage     string            `json:"oauth_image,omitempty"`
	UseCustomImage bool              `json:"use_cimage,omitempty"`
}

// MarshalJSON returns plain text data.
func (s *AccountSettings) MarshalJSON() ([]byte, error) {
	type Alias AccountSettings
	if s.Player.Auth == entities.Local {
		return json.Marshal(&struct {
			LoginID string `json:"email"`
			*Alias
		}{
			LoginID: auther.Decrypt(s.Player.LoginID),
			Alias:   (*Alias)(s),
		})
	}
	return json.Marshal(&struct {
		OAuthName      string `json:"oauth_name"`
		UseCustomName  bool   `json:"use_cname"`
		OAuthImage     string `json:"oauth_image"`
		UseCustomImage bool   `json:"use_cimage"`
		*Alias
	}{
		OAuthName:      auther.Decrypt(s.Player.OAuthDisplayName),
		UseCustomName:  s.Player.UseCustomDisplayName,
		OAuthImage:     auther.Decrypt(s.Player.OAuthImage),
		UseCustomImage: s.Player.UseCustomImage,
		Alias:          (*Alias)(s),
	})
}

// GetAccountSettings returns the list of customizable attributes.
func GetAccountSettings(o *entities.Player) *AccountSettings {
	return &AccountSettings{
		Player:      o,
		CustomName:  auther.Decrypt(o.CustomDisplayName),
		CustomImage: auther.Decrypt(o.CustomImage),
		AuthType:    o.Auth,
	}
}
