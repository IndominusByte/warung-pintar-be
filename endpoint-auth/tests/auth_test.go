package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/go-playground/assert/v2"
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

func TestClearDb(t *testing.T) {
	repo, _ := setupEnvironment()

	user, _ := repo.authRepo.GetUserByEmail(context.Background(), email)
	userConfirm, _ := repo.authRepo.GetUserConfirmByUserId(context.Background(), user.Id)

	repo.authRepo.DeleteUserConfirm(context.Background(), userConfirm.Id)
	repo.authRepo.DeleteUser(context.Background(), user.Id)

	user, _ = repo.authRepo.GetUserByEmail(context.Background(), email_2)
	userConfirm, _ = repo.authRepo.GetUserConfirmByUserId(context.Background(), user.Id)

	repo.authRepo.DeleteUserConfirm(context.Background(), userConfirm.Id)
	repo.authRepo.DeleteUser(context.Background(), user.Id)
}
