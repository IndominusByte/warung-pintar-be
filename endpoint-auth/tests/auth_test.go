package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/IndominusByte/magicimage"
	"github.com/IndominusByte/warung-pintar-be/endpoint-auth/internal/config"
	"github.com/creent-production/cdk-go/auth"
	"github.com/go-chi/jwtauth"
	"github.com/stretchr/testify/assert"
)

const (
	prefix  = "/auth"
	email   = "testtestingtest@test.com"
	email_2 = "testtestingtest2@test.com"
)

func TestValidationRegister(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name    string
		payload map[string]string
	}{

		{
			name:    "required",
			payload: map[string]string{"email": "", "password": "", "confirm_password": ""},
		},
		{
			name:    "minimum",
			payload: map[string]string{"email": "a", "password": "a", "confirm_password": "a"},
		},
		{
			name:    "maximum",
			payload: map[string]string{"email": createMaximum(200), "password": createMaximum(200), "confirm_password": createMaximum(200)},
		},
		{
			name:    "invalid email",
			payload: map[string]string{"email": "test@asdcom"},
		},
		{
			name:    "password not same",
			payload: map[string]string{"password": "asdasd", "confirm_password": "asdqwe"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefix+"/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["email"].(string))
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["password"].(string))
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["confirm_password"].(string))
			case "minimum":
				assert.Equal(t, "Shorter than minimum length 6.", data["detail_message"].(map[string]interface{})["password"].(string))
				assert.Equal(t, "Shorter than minimum length 6.", data["detail_message"].(map[string]interface{})["confirm_password"].(string))
			case "maximum":
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["password"].(string))
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["confirm_password"].(string))
			case "invalid email":
				assert.Equal(t, "Not a valid email address.", data["detail_message"].(map[string]interface{})["email"].(string))
			case "password not same":
				assert.Equal(t, "Must be equal to password.", data["detail_message"].(map[string]interface{})["confirm_password"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestRegister(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name       string
		expected   string
		payload    map[string]string
		statusCode int
	}{
		{
			name:       "create user",
			expected:   "Check your email to activated user.",
			payload:    map[string]string{"email": email, "password": "asdasd", "confirm_password": "asdasd"},
			statusCode: 201,
		},
		{
			name:       "create user 2",
			expected:   "Check your email to activated user.",
			payload:    map[string]string{"email": email_2, "password": "asdasd", "confirm_password": "asdasd"},
			statusCode: 201,
		},
		{
			name:       "duplicate email",
			expected:   "The email has already been taken.",
			payload:    map[string]string{"email": email, "password": "asdasd", "confirm_password": "asdasd"},
			statusCode: 422,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefix+"/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestUserConfirm(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	user, _ := repo.authRepo.GetUserByEmail(context.Background(), email)
	userConfirm, _ := repo.authRepo.GetUserConfirmByUserId(context.Background(), user.Id)

	tests := [...]struct {
		name string
		url  string
	}{
		{
			name: "not found",
			url:  prefix + "/confirm/" + "asd",
		},
		{
			name: "success",
			url:  prefix + "/confirm/" + userConfirm.Id,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, test.url, nil)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "not found":
				assert.Equal(t, "Token not found.", data["detail_message"].(map[string]interface{})["_app"].(string))
				assert.Equal(t, 404, response.Result().StatusCode)
			case "success":
				assert.NotNil(t, data["results"].(map[string]interface{})["access_token"].(string))
				assert.NotNil(t, data["results"].(map[string]interface{})["refresh_token"].(string))
				assert.Equal(t, 200, response.Result().StatusCode)
			}
		})
	}
}

func TestValidationResendEmail(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name    string
		payload map[string]string
	}{

		{
			name:    "required",
			payload: map[string]string{"email": ""},
		},
		{
			name:    "invalid email",
			payload: map[string]string{"email": "test@asdcom"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefix+"/resend-email", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["email"].(string))
			case "invalid email":
				assert.Equal(t, "Not a valid email address.", data["detail_message"].(map[string]interface{})["email"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestResendEmail(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	// decrease delay resend email
	user, _ := repo.authRepo.GetUserByEmail(context.Background(), email_2)
	userConfirm, _ := repo.authRepo.GetUserConfirmByUserId(context.Background(), user.Id)

	repo.authRepo.ResetUserConfirmResendExpired(context.Background(), userConfirm.Id)

	tests := [...]struct {
		name       string
		expected   string
		payload    map[string]string
		statusCode int
	}{
		{
			name:       "not found",
			expected:   "Email not found.",
			payload:    map[string]string{"email": "asdtesting@test.com"},
			statusCode: 404,
		},
		{
			name:       "already activated",
			expected:   "Your account already activated.",
			payload:    map[string]string{"email": email},
			statusCode: 400,
		},
		{
			name:       "success",
			expected:   "Email confirmation has send.",
			payload:    map[string]string{"email": email_2},
			statusCode: 200,
		},
		{
			name:       "delay",
			expected:   "You can try 5 minute later.",
			payload:    map[string]string{"email": email_2},
			statusCode: 400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefix+"/resend-email", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestValidationLogin(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name    string
		payload map[string]string
	}{

		{
			name:    "required",
			payload: map[string]string{"email": "", "password": ""},
		},
		{
			name:    "minimum",
			payload: map[string]string{"email": "a", "password": "a"},
		},
		{
			name:    "maximum",
			payload: map[string]string{"email": createMaximum(200), "password": createMaximum(200)},
		},
		{
			name:    "invalid email",
			payload: map[string]string{"email": "test@asdcom"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["email"].(string))
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["password"].(string))
			case "minimum":
				assert.Equal(t, "Shorter than minimum length 6.", data["detail_message"].(map[string]interface{})["password"].(string))
			case "maximum":
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["password"].(string))
			case "invalid email":
				assert.Equal(t, "Not a valid email address.", data["detail_message"].(map[string]interface{})["email"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestLogin(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name       string
		expected   string
		payload    map[string]string
		statusCode int
	}{
		{
			name:       "email not found",
			expected:   "Invalid credential.",
			payload:    map[string]string{"email": "test@test.com", "password": "asdasd"},
			statusCode: 422,
		},
		{
			name:       "password wrong",
			expected:   "Invalid credential.",
			payload:    map[string]string{"email": email, "password": "asdasd2"},
			statusCode: 422,
		},
		{
			name:       "not activated",
			expected:   "Please check your email to activate your account.",
			payload:    map[string]string{"email": email_2, "password": "asdasd"},
			statusCode: 400,
		},
		{
			name:       "success",
			expected:   "",
			payload:    map[string]string{"email": email, "password": "asdasd"},
			statusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			if test.name != "success" {
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
				assert.Equal(t, test.statusCode, response.Result().StatusCode)
			} else {
				assert.NotNil(t, data["results"].(map[string]interface{})["access_token"].(string))
				assert.NotNil(t, data["results"].(map[string]interface{})["refresh_token"].(string))
				assert.Equal(t, 200, response.Result().StatusCode)
			}
		})
	}
}

func TestValidationFreshToken(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	// login
	body, err := json.Marshal(map[string]string{"email": email, "password": "asdasd"})
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req, s)

	body, _ = io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	accessToken := data["results"].(map[string]interface{})["access_token"].(string)

	tests := [...]struct {
		name    string
		payload map[string]string
	}{
		{
			name:    "required",
			payload: map[string]string{"password": ""},
		},
		{
			name:    "minimum",
			payload: map[string]string{"password": "a"},
		},
		{
			name:    "maximum",
			payload: map[string]string{"password": createMaximum(200)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefix+"/fresh-token", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+accessToken)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["password"].(string))
			case "minimum":
				assert.Equal(t, "Shorter than minimum length 6.", data["detail_message"].(map[string]interface{})["password"].(string))
			case "maximum":
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["password"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestFreshToken(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	// login
	body, err := json.Marshal(map[string]string{"email": email, "password": "asdasd"})
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req, s)

	body, _ = io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	accessToken := data["results"].(map[string]interface{})["access_token"].(string)

	cfg, _ := config.New()
	token := auth.GenerateAccessToken(&auth.AccessToken{Sub: strconv.Itoa(0), Exp: jwtauth.ExpireIn(cfg.JWT.AccessExpires), Fresh: true})
	tokenUserNotFound := auth.NewJwtTokenRSA(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, cfg.JWT.Algorithm, token)

	tests := [...]struct {
		name       string
		payload    map[string]string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			payload:    map[string]string{"password": "asdasd"},
			token:      tokenUserNotFound,
			statusCode: 401,
		},
		{
			name:       "password not same",
			payload:    map[string]string{"password": "asdasd2"},
			token:      accessToken,
			statusCode: 422,
		},
		{
			name:       "success",
			payload:    map[string]string{"password": "asdasd"},
			token:      accessToken,
			statusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefix+"/fresh-token", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+test.token)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found":
				assert.Equal(t, "User not found.", data["detail_message"].(map[string]interface{})["_header"].(string))
			case "password not same":
				assert.Equal(t, "Password does not match with our records.", data["detail_message"].(map[string]interface{})["_app"].(string))
			case "success":
				assert.NotNil(t, data["results"].(map[string]interface{})["access_token"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestRefreshToken(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	// login
	body, err := json.Marshal(map[string]string{"email": email, "password": "asdasd"})
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req, s)

	body, _ = io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	refreshToken := data["results"].(map[string]interface{})["refresh_token"].(string)

	cfg, _ := config.New()
	token := auth.GenerateRefreshToken(&auth.RefreshToken{Sub: strconv.Itoa(0), Exp: jwtauth.ExpireIn(cfg.JWT.RefreshExpires)})
	tokenUserNotFound := auth.NewJwtTokenRSA(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, cfg.JWT.Algorithm, token)

	tests := [...]struct {
		name       string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			token:      tokenUserNotFound,
			statusCode: 401,
		},
		{
			name:       "success",
			token:      refreshToken,
			statusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPost, prefix+"/refresh-token", nil)
			req.Header.Add("Authorization", "Bearer "+test.token)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found":
				assert.Equal(t, "User not found.", data["detail_message"].(map[string]interface{})["_header"].(string))
			case "success":
				assert.NotNil(t, data["results"].(map[string]interface{})["access_token"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestAccessRevoke(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	// login
	body, err := json.Marshal(map[string]string{"email": email, "password": "asdasd"})
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req, s)

	body, _ = io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	accessToken := data["results"].(map[string]interface{})["access_token"].(string)

	tests := [...]struct {
		name       string
		statusCode int
	}{
		{
			name:       "success",
			statusCode: 200,
		},
		{
			name:       "revoked",
			statusCode: 401,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodDelete, prefix+"/access-revoke", nil)
			req.Header.Add("Authorization", "Bearer "+accessToken)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "success":
				assert.Equal(t, "An access token has revoked.", data["detail_message"].(map[string]interface{})["_app"].(string))
			case "revoked":
				assert.Equal(t, "token is revoked", data["detail_message"].(map[string]interface{})["_header"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}

}

func TestRefreshRevoke(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	// login
	body, err := json.Marshal(map[string]string{"email": email, "password": "asdasd"})
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req, s)

	body, _ = io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	refreshToken := data["results"].(map[string]interface{})["refresh_token"].(string)

	tests := [...]struct {
		name       string
		statusCode int
	}{
		{
			name:       "success",
			statusCode: 200,
		},
		{
			name:       "revoked",
			statusCode: 401,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodDelete, prefix+"/refresh-revoke", nil)
			req.Header.Add("Authorization", "Bearer "+refreshToken)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "success":
				assert.Equal(t, "An refresh token has revoked.", data["detail_message"].(map[string]interface{})["_app"].(string))
			case "revoked":
				assert.Equal(t, "token is revoked", data["detail_message"].(map[string]interface{})["_header"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}

}

func TestValidationPasswordResetSend(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name    string
		payload map[string]string
	}{

		{
			name:    "required",
			payload: map[string]string{"email": ""},
		},
		{
			name:    "invalid email",
			payload: map[string]string{"email": "test@asdcom"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefix+"/password-reset/send", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["email"].(string))
			case "invalid email":
				assert.Equal(t, "Not a valid email address.", data["detail_message"].(map[string]interface{})["email"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestPasswordResetSend(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	// decrease delay resend email
	user, _ := repo.authRepo.GetUserByEmail(context.Background(), email)
	userConfirm, _ := repo.authRepo.GetUserConfirmByUserId(context.Background(), user.Id)

	repo.authRepo.ResetUserConfirmResendExpired(context.Background(), userConfirm.Id)

	tests := [...]struct {
		name       string
		expected   string
		payload    map[string]string
		statusCode int
	}{
		{
			name:       "email not found",
			expected:   "We can't find a user with that e-mail address.",
			payload:    map[string]string{"email": "testlol@test.com"},
			statusCode: 404,
		},
		{
			name:       "activate account first",
			expected:   "Please activate your account first.",
			payload:    map[string]string{"email": email_2},
			statusCode: 400,
		},
		{
			name:       "success",
			expected:   "We have sent a password reset link to your email.",
			payload:    map[string]string{"email": email},
			statusCode: 200,
		},
		{
			name:       "delay",
			expected:   "You can try 5 minute later.",
			payload:    map[string]string{"email": email},
			statusCode: 400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefix+"/password-reset/send", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestValidationPasswordReset(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name    string
		payload map[string]string
	}{

		{
			name:    "required",
			payload: map[string]string{"email": "", "password": "", "confirm_password": ""},
		},
		{
			name:    "minimum",
			payload: map[string]string{"email": "a", "password": "a", "confirm_password": "a"},
		},
		{
			name:    "maximum",
			payload: map[string]string{"email": createMaximum(200), "password": createMaximum(200), "confirm_password": createMaximum(200)},
		},
		{
			name:    "invalid email",
			payload: map[string]string{"email": "test@asdcom"},
		},
		{
			name:    "password not same",
			payload: map[string]string{"password": "asdasd", "confirm_password": "asdqwe"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPut, prefix+"/password-reset/token", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["email"].(string))
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["password"].(string))
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["confirm_password"].(string))
			case "minimum":
				assert.Equal(t, "Shorter than minimum length 6.", data["detail_message"].(map[string]interface{})["password"].(string))
				assert.Equal(t, "Shorter than minimum length 6.", data["detail_message"].(map[string]interface{})["confirm_password"].(string))
			case "maximum":
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["password"].(string))
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["confirm_password"].(string))
			case "invalid email":
				assert.Equal(t, "Not a valid email address.", data["detail_message"].(map[string]interface{})["email"].(string))
			case "password not same":
				assert.Equal(t, "Must be equal to password.", data["detail_message"].(map[string]interface{})["confirm_password"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestPasswordReset(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	passwordReset, _ := repo.authRepo.GetPasswordResetByEmail(context.Background(), email)

	tests := [...]struct {
		name       string
		expected   string
		url        string
		payload    map[string]string
		statusCode int
	}{
		{
			name:       "token not found",
			expected:   "Token not found.",
			url:        prefix + "/password-reset/token",
			payload:    map[string]string{"email": email, "password": "asdasd", "confirm_password": "asdasd"},
			statusCode: 404,
		},
		{
			name:       "email not found",
			expected:   "We can't find a user with that e-mail address.",
			url:        prefix + "/password-reset/" + passwordReset.Id,
			payload:    map[string]string{"email": "testlol@test.com", "password": "asdasd", "confirm_password": "asdasd"},
			statusCode: 404,
		},
		{
			name:       "token not same as email",
			expected:   "The password reset token is invalid.",
			url:        prefix + "/password-reset/" + passwordReset.Id,
			payload:    map[string]string{"email": email_2, "password": "asdasd", "confirm_password": "asdasd"},
			statusCode: 400,
		},
		{
			name:       "success",
			expected:   "Successfully reset your password.",
			url:        prefix + "/password-reset/" + passwordReset.Id,
			payload:    map[string]string{"email": email, "password": "asdasd", "confirm_password": "asdasd"},
			statusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPut, test.url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestValidationUpdatePassword(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	// login
	body, err := json.Marshal(map[string]string{"email": email, "password": "asdasd"})
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req, s)

	body, _ = io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	accessToken := data["results"].(map[string]interface{})["access_token"].(string)

	tests := [...]struct {
		name    string
		payload map[string]string
	}{
		{
			name:    "required",
			payload: map[string]string{"old_password": "", "password": "", "confirm_password": ""},
		},
		{
			name:    "minimum",
			payload: map[string]string{"old_password": "a", "password": "a", "confirm_password": "a"},
		},
		{
			name:    "maximum",
			payload: map[string]string{"old_password": createMaximum(200), "password": createMaximum(200), "confirm_password": createMaximum(200)},
		},
		{
			name:    "password not same",
			payload: map[string]string{"password": "asdasd", "confirm_password": "asdqwe"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPut, prefix+"/update-password", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+accessToken)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["old_password"].(string))
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["password"].(string))
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["confirm_password"].(string))
			case "minimum":
				assert.Equal(t, "Shorter than minimum length 6.", data["detail_message"].(map[string]interface{})["old_password"].(string))
				assert.Equal(t, "Shorter than minimum length 6.", data["detail_message"].(map[string]interface{})["password"].(string))
				assert.Equal(t, "Shorter than minimum length 6.", data["detail_message"].(map[string]interface{})["confirm_password"].(string))
			case "maximum":
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["old_password"].(string))
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["password"].(string))
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["confirm_password"].(string))
			case "password not same":
				assert.Equal(t, "Must be equal to password.", data["detail_message"].(map[string]interface{})["confirm_password"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestUpdatePassword(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	// login
	body, err := json.Marshal(map[string]string{"email": email, "password": "asdasd"})
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req, s)

	body, _ = io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	accessToken := data["results"].(map[string]interface{})["access_token"].(string)

	cfg, _ := config.New()
	token := auth.GenerateAccessToken(&auth.AccessToken{Sub: strconv.Itoa(0), Exp: jwtauth.ExpireIn(cfg.JWT.AccessExpires), Fresh: true})
	tokenUserNotFound := auth.NewJwtTokenRSA(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, cfg.JWT.Algorithm, token)

	tests := [...]struct {
		name       string
		payload    map[string]string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			payload:    map[string]string{"old_password": "asdasd", "password": "asdasd", "confirm_password": "asdasd"},
			token:      tokenUserNotFound,
			statusCode: 401,
		},
		{
			name:       "old password wrong",
			payload:    map[string]string{"old_password": "asdasdasd", "password": "asdasd", "confirm_password": "asdasd"},
			token:      accessToken,
			statusCode: 400,
		},
		{
			name:       "success",
			payload:    map[string]string{"old_password": "asdasd", "password": "asdasd", "confirm_password": "asdasd"},
			token:      accessToken,
			statusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPut, prefix+"/update-password", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+test.token)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found":
				assert.Equal(t, "User not found.", data["detail_message"].(map[string]interface{})["_header"].(string))
			case "old password wrong":
				assert.Equal(t, "Password does not match with our records.", data["detail_message"].(map[string]interface{})["_app"].(string))
			case "success":
				assert.Equal(t, "Success update your password.", data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)

		})
	}
}

func TestValidationUpdateAvatar(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	// login
	body, err := json.Marshal(map[string]string{"email": email, "password": "asdasd"})
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req, s)

	body, _ = io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	accessToken := data["results"].(map[string]interface{})["access_token"].(string)

	tests := [...]struct {
		name     string
		expected string
		form     map[string]string
	}{
		{
			name:     "required",
			expected: "Image is required.",
			form:     map[string]string{},
		},
		{
			name:     "danger file extension",
			expected: "Image must be between jpeg, png.",
			form:     map[string]string{"avatar": "@/app/static/test_image/test.txt"},
		},
		{
			name:     "not valid file extension",
			expected: "Image must be between jpeg, png.",
			form:     map[string]string{"avatar": "@/app/static/test_image/test.gif"},
		},
		{
			name:     "file cannot grater than 4 Mb",
			expected: "An image cannot greater than 4 Mb.",
			form:     map[string]string{"avatar": "@/app/static/test_image/size.png"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ct, b, err := createForm(test.form)
			if err != nil {
				panic(err)
			}

			req, _ = http.NewRequest(http.MethodPut, prefix+"/update-avatar", b)
			req.Header.Add("Authorization", "Bearer "+accessToken)
			req.Header.Set("Content-Type", ct)

			response = executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["avatar"].(string))
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestUpdateAvatar(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	// login
	body, err := json.Marshal(map[string]string{"email": email, "password": "asdasd"})
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req, s)

	body, _ = io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	accessToken := data["results"].(map[string]interface{})["access_token"].(string)

	cfg, _ := config.New()
	token := auth.GenerateAccessToken(&auth.AccessToken{Sub: strconv.Itoa(0), Exp: jwtauth.ExpireIn(cfg.JWT.AccessExpires), Fresh: true})
	tokenUserNotFound := auth.NewJwtTokenRSA(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, cfg.JWT.Algorithm, token)

	tests := [...]struct {
		name       string
		payload    map[string]string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			payload:    map[string]string{"avatar": "@/app/static/test_image/image.jpeg"},
			token:      tokenUserNotFound,
			statusCode: 401,
		},
		{
			name:       "success",
			payload:    map[string]string{"avatar": "@/app/static/test_image/image.jpeg"},
			token:      accessToken,
			statusCode: 200,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ct, b, err := createForm(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ = http.NewRequest(http.MethodPut, prefix+"/update-avatar", b)
			req.Header.Add("Authorization", "Bearer "+test.token)
			req.Header.Set("Content-Type", ct)

			response = executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found":
				assert.Equal(t, "User not found.", data["detail_message"].(map[string]interface{})["_header"].(string))
			case "success":
				assert.Equal(t, "Success update avatar.", data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestValidationUpdateAccount(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	// login
	body, err := json.Marshal(map[string]string{"email": email, "password": "asdasd"})
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req, s)

	body, _ = io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	accessToken := data["results"].(map[string]interface{})["access_token"].(string)

	tests := [...]struct {
		name    string
		payload map[string]string
	}{
		{
			name:    "required",
			payload: map[string]string{"fullname": "", "phone": "", "address": ""},
		},
		{
			name:    "minimum",
			payload: map[string]string{"fullname": "a", "phone": "a", "address": "a"},
		},
		{
			name:    "maximum",
			payload: map[string]string{"fullname": createMaximum(200), "phone": createMaximum(200), "address": createMaximum(200)},
		},
		{
			name:    "invalid phone",
			payload: map[string]string{"phone": "87862265363"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPut, prefix+"/update-account", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+accessToken)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["fullname"].(string))
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["phone"].(string))
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["address"].(string))
			case "minimum":
				assert.Equal(t, "Shorter than minimum length 3.", data["detail_message"].(map[string]interface{})["fullname"].(string))
				assert.Equal(t, "Shorter than minimum length 5.", data["detail_message"].(map[string]interface{})["address"].(string))
			case "maximum":
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["fullname"].(string))
			case "invalid phone":
				assert.Equal(t, "Invalid phone number.", data["detail_message"].(map[string]interface{})["phone"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestUpdateAccount(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}
	// set activated
	user, _ := repo.authRepo.GetUserByEmail(context.Background(), email_2)
	userConfirm, _ := repo.authRepo.GetUserConfirmByUserId(context.Background(), user.Id)
	repo.authRepo.SetUserConfirmActivatedTrue(context.Background(), userConfirm.Id)

	// login
	body, err := json.Marshal(map[string]string{"email": email, "password": "asdasd"})
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req, s)

	body, _ = io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	accessToken := data["results"].(map[string]interface{})["access_token"].(string)

	// login user 2
	body, err = json.Marshal(map[string]string{"email": email_2, "password": "asdasd"})
	if err != nil {
		panic(err)
	}

	req, _ = http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	response = executeRequest(req, s)

	body, _ = io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	accessToken2 := data["results"].(map[string]interface{})["access_token"].(string)

	cfg, _ := config.New()
	token := auth.GenerateAccessToken(&auth.AccessToken{Sub: strconv.Itoa(0), Exp: jwtauth.ExpireIn(cfg.JWT.AccessExpires), Fresh: true})
	tokenUserNotFound := auth.NewJwtTokenRSA(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, cfg.JWT.Algorithm, token)

	tests := [...]struct {
		name       string
		payload    map[string]string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			payload:    map[string]string{"fullname": "oman", "phone": "087862265363", "address": "testing"},
			token:      tokenUserNotFound,
			statusCode: 401,
		},
		{
			name:       "success",
			payload:    map[string]string{"fullname": "oman", "phone": "087862265363", "address": "testing"},
			token:      accessToken,
			statusCode: 200,
		},
		{
			name:       "phone taken",
			payload:    map[string]string{"fullname": "oman", "phone": "087862265363", "address": "testing"},
			token:      accessToken2,
			statusCode: 400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPut, prefix+"/update-account", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+test.token)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found":
				assert.Equal(t, "User not found.", data["detail_message"].(map[string]interface{})["_header"].(string))
			case "success":
				assert.Equal(t, "Success updated your account.", data["detail_message"].(map[string]interface{})["_app"].(string))
			case "phone taken":
				assert.Equal(t, "The phone has already been taken.", data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestMyData(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}
	// login
	body, err := json.Marshal(map[string]string{"email": email, "password": "asdasd"})
	if err != nil {
		panic(err)
	}

	req, _ := http.NewRequest(http.MethodPost, prefix+"/login", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	response := executeRequest(req, s)

	body, _ = io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	accessToken := data["results"].(map[string]interface{})["access_token"].(string)

	req, _ = http.NewRequest(http.MethodGet, prefix, bytes.NewBuffer(body))
	req.Header.Add("Authorization", "Bearer "+accessToken)

	response = executeRequest(req, s)

	body, _ = io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	assert.NotNil(t, data["results"])
	assert.Equal(t, 200, response.Result().StatusCode)
}

func TestClearDb(t *testing.T) {
	repo, _ := setupEnvironment()

	user, _ := repo.authRepo.GetUserByEmail(context.Background(), email)
	userConfirm, _ := repo.authRepo.GetUserConfirmByUserId(context.Background(), user.Id)

	repo.authRepo.DeleteUserConfirm(context.Background(), userConfirm.Id)
	repo.authRepo.DeleteUser(context.Background(), user.Id)
	if user.Avatar != "default.jpg" {
		magicimage.DeleteFolderAndFile(fmt.Sprintf("/app/static/avatars/%s", user.Avatar))
	}

	user, _ = repo.authRepo.GetUserByEmail(context.Background(), email_2)
	userConfirm, _ = repo.authRepo.GetUserConfirmByUserId(context.Background(), user.Id)

	repo.authRepo.DeleteUserConfirm(context.Background(), userConfirm.Id)
	repo.authRepo.DeleteUser(context.Background(), user.Id)
	if user.Avatar != "default.jpg" {
		magicimage.DeleteFolderAndFile(fmt.Sprintf("/app/static/avatars/%s", user.Avatar))
	}
}
