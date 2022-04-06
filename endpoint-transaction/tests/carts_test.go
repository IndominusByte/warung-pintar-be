package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/config"
	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/auth"
	categoriesentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/categories"
	productsentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/products"
	"github.com/creent-production/cdk-go/auth"
	"github.com/go-chi/jwtauth"
	"github.com/stretchr/testify/assert"
)

const (
	prefixCart = "/carts"
	email      = "testtestingtest@test.com"
	email_2    = "testtestingtest2@test.com"
	namee      = "test"
	namee_2    = "test2"
)

var (
	productId     = 0
	productId2    = 0
	tokenAdmin    = ""
	tokenGuest    = ""
	tokenNotFound = ""
)

func TestUpCart(t *testing.T) {
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

	// create category
	categoryId := repo.categoriesRepo.Insert(context.Background(), &categoriesentity.JsonCreateUpdateSchema{Name: namee})
	categoryId2 := repo.categoriesRepo.Insert(context.Background(), &categoriesentity.JsonCreateUpdateSchema{Name: namee_2})

	// create product
	payload := productsentity.FormCreateUpdateSchema{
		Name:        namee,
		Slug:        namee,
		Description: "asdasd",
		Image:       "asdasd",
		Price:       1,
		Stock:       1,
		CategoryId:  categoryId,
	}
	productId = repo.productsRepo.Insert(context.Background(), &payload)
	payload2 := productsentity.FormCreateUpdateSchema{
		Name:        namee_2,
		Slug:        namee_2,
		Description: "asdasd",
		Image:       "asdasd",
		Price:       1,
		Stock:       2,
		CategoryId:  categoryId2,
	}
	productId2 = repo.productsRepo.Insert(context.Background(), &payload2)
}

func TestValidationPutToCart(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name:    "required",
			payload: map[string]interface{}{},
		},
		{
			name:    "minimum",
			payload: map[string]interface{}{"operation": "a", "product_id": -1, "notes": "a", "qty": -1},
		},
		{
			name:    "maximum",
			payload: map[string]interface{}{"notes": createMaximum(200)},
		},
		{
			name:    "one of",
			payload: map[string]interface{}{"operation": "a"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefixCart+"/put-product", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+tokenAdmin)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["operation"].(string))
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["product_id"].(string))
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["qty"].(string))
			case "minimum":
				assert.Equal(t, "Shorter than minimum length 3.", data["detail_message"].(map[string]interface{})["notes"].(string))
				assert.Equal(t, "Must be greater than or equal to 1.", data["detail_message"].(map[string]interface{})["product_id"].(string))
				assert.Equal(t, "Must be greater than or equal to 1.", data["detail_message"].(map[string]interface{})["qty"].(string))
			case "maximum":
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["notes"].(string))
			case "one of":
				assert.Equal(t, "Must be one of: create, update.", data["detail_message"].(map[string]interface{})["operation"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestPutToCart(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name       string
		payload    map[string]interface{}
		expected   string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			payload:    map[string]interface{}{"operation": "create", "product_id": productId, "notes": "asdasd", "qty": 1},
			expected:   "User not found.",
			token:      tokenNotFound,
			statusCode: 401,
		},
		{
			name:       "product not found",
			payload:    map[string]interface{}{"operation": "create", "product_id": 9999999, "notes": "asdasd", "qty": 1},
			expected:   "Product not found.",
			token:      tokenGuest,
			statusCode: 404,
		},
		{
			name:       "exceed stock",
			payload:    map[string]interface{}{"operation": "create", "product_id": productId, "notes": "asdasd", "qty": 10},
			expected:   "The amount you input exceeds the available stock.",
			token:      tokenGuest,
			statusCode: 400,
		},
		{
			name:       "success create",
			payload:    map[string]interface{}{"operation": "create", "product_id": productId, "notes": "asdasd", "qty": 1},
			expected:   "The product has been successfully added to the shopping cart.",
			token:      tokenGuest,
			statusCode: 201,
		},
		{
			name:       "exceed stock update",
			payload:    map[string]interface{}{"operation": "update", "product_id": productId, "notes": "asdasd", "qty": 10},
			expected:   "This item has 1 stock left and you already have 1 in your basket.",
			token:      tokenGuest,
			statusCode: 400,
		},
		{
			name:       "success update",
			payload:    map[string]interface{}{"operation": "update", "product_id": productId, "qty": 1},
			expected:   "Shopping cart successfully updated.",
			token:      tokenGuest,
			statusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefixCart+"/put-product", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+test.token)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_header"].(string))
			default:
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestValidationGetAllDataCart(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name string
		url  string
	}{
		{
			name: "one of",
			url:  prefixCart + "?stock=a",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, test.url, nil)
			req.Header.Add("Authorization", "Bearer "+tokenGuest)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "one of":
				assert.Equal(t, "Must be one of: empty, ready.", data["detail_message"].(map[string]interface{})["stock"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestGetAllDataCart(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	req, _ := http.NewRequest(http.MethodGet, prefixCart, nil)
	req.Header.Add("Authorization", "Bearer "+tokenGuest)

	response := executeRequest(req, s)

	body, _ := io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	assert.NotNil(t, data["results"])
	assert.Equal(t, 200, response.Result().StatusCode)
}

func TestValidationDeleteCart(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name:    "required",
			payload: map[string]interface{}{},
		},
		{
			name:    "unique",
			payload: map[string]interface{}{"list_id": []int{1, 1}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodDelete, prefixCart, bytes.NewBuffer(body))
			req.Header.Add("Authorization", "Bearer "+tokenGuest)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["list_id"].(string))
			case "unique":
				assert.Equal(t, "Must be unique.", data["detail_message"].(map[string]interface{})["list_id"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestDeleteCart(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name       string
		payload    map[string]interface{}
		expected   string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			payload:    map[string]interface{}{"list_id": []int{9999, 9998}},
			expected:   "User not found.",
			token:      tokenNotFound,
			statusCode: 401,
		},
		{
			name:       "success",
			payload:    map[string]interface{}{"list_id": []int{9999, 9998}},
			expected:   "2 items were removed.",
			token:      tokenGuest,
			statusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodDelete, prefixCart, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+test.token)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_header"].(string))
			default:
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestValidationMoveToPaymentCart(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name    string
		payload map[string]interface{}
	}{
		{
			name:    "required",
			payload: map[string]interface{}{},
		},
		{
			name:    "unique",
			payload: map[string]interface{}{"list_id": []int{1, 1}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefixCart+"/move-to-payment", bytes.NewBuffer(body))
			req.Header.Add("Authorization", "Bearer "+tokenGuest)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["list_id"].(string))
			case "unique":
				assert.Equal(t, "Must be unique.", data["detail_message"].(map[string]interface{})["list_id"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestMoveToPaymentCart(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	user, _ := repo.authRepo.GetUserByEmail(context.Background(), email_2)
	cart, _ := repo.cartsRepo.GetCartByUserIdAndProductId(context.Background(), user.Id, productId)

	tests := [...]struct {
		name       string
		payload    map[string]interface{}
		expected   string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			payload:    map[string]interface{}{"list_id": []int{9999, 9998}},
			expected:   "User not found.",
			token:      tokenNotFound,
			statusCode: 401,
		},
		{
			name:       "success",
			payload:    map[string]interface{}{"list_id": []int{cart.Id, 9998}},
			expected:   "2 items successfully moved to the payment.",
			token:      tokenGuest,
			statusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefixCart+"/move-to-payment", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Add("Authorization", "Bearer "+test.token)

			response := executeRequest(req, s)

			body, _ = io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_header"].(string))
			default:
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestGetAllItemInPaymentCart(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	req, _ := http.NewRequest(http.MethodGet, prefixCart+"/item-in-payment", nil)
	req.Header.Add("Authorization", "Bearer "+tokenGuest)

	response := executeRequest(req, s)

	body, _ := io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	assert.NotNil(t, data["results"])
	assert.Equal(t, 200, response.Result().StatusCode)
}
