package recaptcha

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	kklogger "github.com/yetiz-org/goth-kklogger"
	"github.com/yetiz-org/goth-scaffold/app/conf"
)

// Response represents the response from Google reCAPTCHA API
type Response struct {
	Success bool    `json:"success"`
	Score   float64 `json:"score,omitempty"`  // v3 only
	Action  string  `json:"action,omitempty"` // v3 only
}

// Verify verifies recaptcha token with Google reCAPTCHA API
func Verify(token string) bool {
	secretKey := conf.Config().Credentials.Recaptcha.SecretKey
	verifyUrl := "https://www.google.com/recaptcha/api/siteverify"

	params := url.Values{
		"secret":   {secretKey},
		"response": {token},
	}

	httpResp, err := http.PostForm(verifyUrl, params)
	if err != nil {
		kklogger.ErrorJ("recaptcha:Verify#google!request_fail", err.Error())
		return false
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		kklogger.ErrorJ("recaptcha:Verify#google!response_read", err.Error())
		return false
	}

	var data Response
	if err := json.Unmarshal(body, &data); err != nil {
		kklogger.ErrorJ("recaptcha:Verify#google!response_parse", err.Error())
		return false
	}

	return data.Success
}
