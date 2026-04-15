package googleauth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	buf "github.com/yetiz-org/goth-bytebuf"
	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/components/httpclient"
)

// TokenResponse represents Google OAuth token response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	Scope        string `json:"scope"`
}

// UserInfo represents Google user information.
type UserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"`
}

// Client holds Google OAuth credentials and provides authentication helpers.
type Client struct {
	ClientId     string
	ClientSecret string
}

// NewClient creates a new googleauth Client with the given credentials.
func NewClient(clientId, clientSecret string) *Client {
	return &Client{
		ClientId:     clientId,
		ClientSecret: clientSecret,
	}
}

// GenerateOAuthUrl builds the Google OAuth 2.0 authorization URL.
func (c *Client) GenerateOAuthUrl(redirectUri, state string) string {
	baseUrl, _ := url.Parse("https://accounts.google.com/o/oauth2/v2/auth")
	params := url.Values{
		"client_id":     {c.ClientId},
		"redirect_uri":  {redirectUri},
		"response_type": {"code"},
		"scope":         {"openid email profile"},
		"state":         {state},
	}

	baseUrl.RawQuery = params.Encode()
	return baseUrl.String()
}

// ExchangeCodeForToken exchanges an authorization code for OAuth tokens.
func (c *Client) ExchangeCodeForToken(code, redirectUri string) (*TokenResponse, error) {
	tokenURL := "https://oauth2.googleapis.com/token"

	data := url.Values{}
	data.Set("client_id", c.ClientId)
	data.Set("client_secret", c.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", redirectUri)

	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		kklogger.ErrorJ("googleauth:Client.ExchangeCodeForToken#request!create_fail", err.Error())
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := httpclient.DoAndLog(req)
	if err != nil {
		kklogger.ErrorJ("googleauth:Client.ExchangeCodeForToken#google!request_fail", err.Error())
		return nil, err
	}

	body := resp.Body.(buf.ByteBuf).Bytes()
	if resp.StatusCode != http.StatusOK {
		kklogger.ErrorJ("googleauth:Client.ExchangeCodeForToken#google!token_exchange_fail", string(body))
		return nil, fmt.Errorf("token exchange failed: %s", string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		kklogger.ErrorJ("googleauth:Client.ExchangeCodeForToken#google!response_parse", err.Error())
		return nil, err
	}

	return &tokenResp, nil
}

// GetUserInfo retrieves user information from Google using an access token.
func (c *Client) GetUserInfo(accessToken string) (*UserInfo, error) {
	userInfoURL := "https://www.googleapis.com/oauth2/v2/userinfo"

	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		kklogger.ErrorJ("googleauth:Client.GetUserInfo#request!create_fail", err.Error())
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := httpclient.DoAndLog(req)
	if err != nil {
		kklogger.ErrorJ("googleauth:Client.GetUserInfo#google!request_fail", err.Error())
		return nil, err
	}

	body := resp.Body.(buf.ByteBuf).Bytes()
	if resp.StatusCode != http.StatusOK {
		kklogger.ErrorJ("googleauth:Client.GetUserInfo#google!get_user_info_fail", string(body))
		return nil, fmt.Errorf("get user info failed: %s", string(body))
	}

	var userInfo UserInfo
	if err := json.Unmarshal(body, &userInfo); err != nil {
		kklogger.ErrorJ("googleauth:Client.GetUserInfo#google!response_parse", err.Error())
		return nil, err
	}

	return &userInfo, nil
}
