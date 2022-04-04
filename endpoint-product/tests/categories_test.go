package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/config"
	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/auth"
	"github.com/creent-production/cdk-go/auth"
	"github.com/go-chi/jwtauth"
	"github.com/stretchr/testify/assert"
)

const (
	prefixCategory = "/categories"
	email          = "testtestingtest@test.com"
	email_2        = "testtestingtest2@test.com"
	namee          = "test"
	namee_2        = "test2"
)

var (
	tokenAdmin    = ""
	tokenGuest    = ""
	tokenNotFound = ""
)

func TestUpCategory(t *testing.T) {
	repo, _ := setupEnvironment()
	cfg, _ := config.New()
	// create admin
	user_id := repo.authRepo.InsertUser(context.Background(), &authentity.JsonRegisterSchema{Email: email, Password: "asdasd"})
	repo.authRepo.InsertUserConfirm(context.Background(), user_id)
	repo.authRepo.SetUserConfirmActivatedTrue(context.Background(), user_id)
	repo.authRepo.SetUserRoleAdmin(context.Background(), email)

	token := auth.GenerateAccessToken(&auth.AccessToken{Sub: strconv.Itoa(user_id), Exp: jwtauth.ExpireIn(cfg.JWT.AccessExpires), Fresh: true})
	tokenAdmin = auth.NewJwtTokenRSA(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, cfg.JWT.Algorithm, token)

	// crease guest
	user_id = repo.authRepo.InsertUser(context.Background(), &authentity.JsonRegisterSchema{Email: email_2, Password: "asdasd"})
	repo.authRepo.InsertUserConfirm(context.Background(), user_id)
	repo.authRepo.SetUserConfirmActivatedTrue(context.Background(), user_id)

	token = auth.GenerateAccessToken(&auth.AccessToken{Sub: strconv.Itoa(user_id), Exp: jwtauth.ExpireIn(cfg.JWT.AccessExpires), Fresh: true})
	tokenGuest = auth.NewJwtTokenRSA(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, cfg.JWT.Algorithm, token)

	token = auth.GenerateAccessToken(&auth.AccessToken{Sub: strconv.Itoa(0), Exp: jwtauth.ExpireIn(cfg.JWT.AccessExpires), Fresh: true})
	tokenNotFound = auth.NewJwtTokenRSA(cfg.JWT.PublicKey, cfg.JWT.PrivateKey, cfg.JWT.Algorithm, token)
}

func TestValidationCreateCategory(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name    string
		payload map[string]string
	}{
		{
			name:    "required",
			payload: map[string]string{"name": ""},
		},
		{
			name:    "minimum",
			payload: map[string]string{"name": "a"},
		},
		{
			name:    "maximum",
			payload: map[string]string{"name": createMaximum(200)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefixCategory, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+tokenAdmin)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["name"].(string))
			case "minimum":
				assert.Equal(t, "Shorter than minimum length 3.", data["detail_message"].(map[string]interface{})["name"].(string))
			case "maximum":
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["name"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}

}

func TestCreateCategory(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name       string
		payload    map[string]string
		expected   string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			payload:    map[string]string{"name": namee},
			expected:   "User not found.",
			token:      tokenNotFound,
			statusCode: 401,
		},
		{
			name:       "user not admin",
			payload:    map[string]string{"name": namee},
			expected:   "Only users with admin privileges can do this action.",
			token:      tokenGuest,
			statusCode: 401,
		},
		{
			name:       "success",
			payload:    map[string]string{"name": namee},
			expected:   "Successfully add a new category.",
			token:      tokenAdmin,
			statusCode: 201,
		},
		{
			name:       "success",
			payload:    map[string]string{"name": namee_2},
			expected:   "Successfully add a new category.",
			token:      tokenAdmin,
			statusCode: 201,
		},
		{
			name:       "name already taken",
			payload:    map[string]string{"name": namee},
			expected:   "The name has already been taken.",
			token:      tokenAdmin,
			statusCode: 400,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefixCategory, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+test.token)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found", "user not admin":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_header"].(string))
			case "success", "name already taken":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestValidationGetAllDataCategory(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name string
		url  string
	}{
		{
			name: "empty",
			url:  prefixCategory,
		},
		{
			name: "required",
			url:  prefixCategory + "?page=&per_page=&q=",
		},
		{
			name: "type data",
			url:  prefixCategory + "?page=a&per_page=a&q=1",
		},
		{
			name: "minimum",
			url:  prefixCategory + "?page=-1&per_page=-1&q=",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, test.url, nil)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "empty", "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["page"].(string))
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["per_page"].(string))
			case "type data":
				assert.Equal(t, "Invalid input type.", data["detail_message"].(map[string]interface{})["_body"].(string))
			case "minimum":
				assert.Equal(t, "Must be greater than or equal to 1.", data["detail_message"].(map[string]interface{})["page"].(string))
				assert.Equal(t, "Must be greater than or equal to 1.", data["detail_message"].(map[string]interface{})["per_page"].(string))
			}

			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestGetAllDataCategory(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	req, _ := http.NewRequest(http.MethodGet, prefixCategory+"?page=1&per_page=1&q=t", nil)

	response := executeRequest(req, s)

	body, _ := io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	assert.NotNil(t, data["results"].(map[string]interface{})["data"])
	assert.Equal(t, 200, response.Result().StatusCode)
}

func TestValidationGetByIdCategory(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name string
		url  string
	}{
		{
			name: "type data",
			url:  prefixCategory + "/abc",
		},
		{
			name: "minimum",
			url:  prefixCategory + "/-1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, test.url, nil)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			assert.Equal(t, "404 page not found", strings.TrimSuffix(string(body), "\n"))
			assert.Equal(t, 404, response.Result().StatusCode)
		})
	}
}

func TestGetByIdCategory(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	// get id
	category, _ := repo.categoriesRepo.GetCategoryByName(context.Background(), namee)

	tests := [...]struct {
		name string
		url  string
	}{
		{
			name: "not found",
			url:  prefixCategory + "/99999999",
		},
		{
			name: "success",
			url:  prefixCategory + "/" + strconv.Itoa(category.Id),
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
				assert.Equal(t, "Category not found.", data["detail_message"].(map[string]interface{})["_app"].(string))
				assert.Equal(t, 404, response.Result().StatusCode)
			case "success":
				assert.NotNil(t, data["results"])
				assert.Equal(t, 200, response.Result().StatusCode)
			}
		})
	}
}

func TestValidationUpdateCategory(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	// get id
	category, _ := repo.categoriesRepo.GetCategoryByName(context.Background(), namee)

	tests := [...]struct {
		name    string
		payload map[string]string
	}{
		{
			name:    "required",
			payload: map[string]string{"name": ""},
		},
		{
			name:    "minimum",
			payload: map[string]string{"name": "a"},
		},
		{
			name:    "maximum",
			payload: map[string]string{"name": createMaximum(200)},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPut, prefixCategory+"/"+strconv.Itoa(category.Id), bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+tokenAdmin)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["name"].(string))
			case "minimum":
				assert.Equal(t, "Shorter than minimum length 3.", data["detail_message"].(map[string]interface{})["name"].(string))
			case "maximum":
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["name"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestUpdateCategory(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	// get id
	category, _ := repo.categoriesRepo.GetCategoryByName(context.Background(), namee)

	tests := [...]struct {
		name       string
		url        string
		payload    map[string]string
		expected   string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			url:        prefixCategory + "/99999999",
			payload:    map[string]string{"name": namee},
			expected:   "User not found.",
			token:      tokenNotFound,
			statusCode: 401,
		},
		{
			name:       "user not admin",
			url:        prefixCategory + "/99999999",
			payload:    map[string]string{"name": namee},
			expected:   "Only users with admin privileges can do this action.",
			token:      tokenGuest,
			statusCode: 401,
		},
		{
			name:       "category not found",
			url:        prefixCategory + "/99999999",
			payload:    map[string]string{"name": namee},
			expected:   "Category not found.",
			token:      tokenAdmin,
			statusCode: 404,
		},
		{
			name:       "name already taken",
			url:        prefixCategory + "/" + strconv.Itoa(category.Id),
			payload:    map[string]string{"name": namee_2},
			expected:   "The name has already been taken.",
			token:      tokenAdmin,
			statusCode: 400,
		},
		{
			name:       "success",
			url:        prefixCategory + "/" + strconv.Itoa(category.Id),
			payload:    map[string]string{"name": namee},
			expected:   "Successfully update the category.",
			token:      tokenAdmin,
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
			req.Header.Add("Authorization", "Bearer "+test.token)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found", "user not admin":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_header"].(string))
			case "category not found", "name already taken", "success":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}

}

func TestValidationDeleteCategory(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name string
		url  string
	}{
		{
			name: "type data",
			url:  prefixCategory + "/abc",
		},
		{
			name: "minimum",
			url:  prefixCategory + "/-1",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodDelete, test.url, nil)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			assert.Equal(t, "404 page not found", strings.TrimSuffix(string(body), "\n"))
			assert.Equal(t, 404, response.Result().StatusCode)
		})
	}
}

func TestDeleteCategory(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	// get id
	category, _ := repo.categoriesRepo.GetCategoryByName(context.Background(), namee)
	category_2, _ := repo.categoriesRepo.GetCategoryByName(context.Background(), namee_2)

	tests := [...]struct {
		name       string
		expected   string
		url        string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			expected:   "User not found.",
			url:        prefixCategory + "/99999999",
			token:      tokenNotFound,
			statusCode: 401,
		},
		{
			name:       "user not admin",
			expected:   "Only users with admin privileges can do this action.",
			url:        prefixCategory + "/99999999",
			token:      tokenGuest,
			statusCode: 401,
		},
		{
			name:       "category not found",
			expected:   "Category not found.",
			url:        prefixCategory + "/99999999",
			token:      tokenAdmin,
			statusCode: 404,
		},
		{
			name:       "success",
			expected:   "Successfully delete the category.",
			url:        prefixCategory + "/" + strconv.Itoa(category.Id),
			token:      tokenAdmin,
			statusCode: 200,
		},
		{
			name:       "success",
			expected:   "Successfully delete the category.",
			url:        prefixCategory + "/" + strconv.Itoa(category_2.Id),
			token:      tokenAdmin,
			statusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodDelete, test.url, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+test.token)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found", "user not admin":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_header"].(string))
			case "category not found", "success":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)

		})
	}
}

func TestDownCategory(t *testing.T) {
	repo, _ := setupEnvironment()

	user, _ := repo.authRepo.GetUserByEmail(context.Background(), email)
	userConfirm, _ := repo.authRepo.GetUserConfirmByUserId(context.Background(), user.Id)

	repo.authRepo.DeleteUserConfirm(context.Background(), userConfirm.Id)
	repo.authRepo.DeleteUser(context.Background(), user.Id)

	user, _ = repo.authRepo.GetUserByEmail(context.Background(), email_2)
	userConfirm, _ = repo.authRepo.GetUserConfirmByUserId(context.Background(), user.Id)

	repo.authRepo.DeleteUserConfirm(context.Background(), userConfirm.Id)
	repo.authRepo.DeleteUser(context.Background(), user.Id)

	category, _ := repo.categoriesRepo.GetCategoryByName(context.Background(), namee)
	repo.categoriesRepo.Delete(context.Background(), category.Id)

	category, _ = repo.categoriesRepo.GetCategoryByName(context.Background(), namee_2)
	repo.categoriesRepo.Delete(context.Background(), category.Id)
}
