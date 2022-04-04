package tests

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/config"
	authentity "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/auth"
	"github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/categories"
	"github.com/creent-production/cdk-go/auth"
	"github.com/go-chi/jwtauth"
	"github.com/stretchr/testify/assert"
)

const (
	prefixProduct = "/products"
)

func TestUpProduct(t *testing.T) {
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
	repo.categoriesRepo.Insert(context.Background(), &categories.JsonCreateUpdateSchema{Name: namee})
	repo.categoriesRepo.Insert(context.Background(), &categories.JsonCreateUpdateSchema{Name: namee_2})
}

func TestValidationCreateProduct(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name    string
		payload map[string]string
	}{
		{
			name:    "minimum",
			payload: map[string]string{"name": "a", "description": "a", "image": "@/app/static/test_image/image.jpeg", "price": "-1", "stock": "-1", "category_id": "-1"},
		},
		{
			name:    "maximum",
			payload: map[string]string{"name": createMaximum(200), "description": createMaximum(200), "image": "@/app/static/test_image/image.jpeg"},
		},
		{
			name:    "danger file extension",
			payload: map[string]string{"image": "@/app/static/test_image/test.txt"},
		},
		{
			name:    "not valid file extension",
			payload: map[string]string{"image": "@/app/static/test_image/test.gif"},
		},
		{
			name:    "file cannot grater than 4 Mb",
			payload: map[string]string{"image": "@/app/static/test_image/size.png"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ct, b, err := createForm(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefixProduct, b)
			req.Header.Add("Authorization", "Bearer "+tokenAdmin)
			req.Header.Set("Content-Type", ct)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "minimum":
				assert.Equal(t, "Shorter than minimum length 3.", data["detail_message"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "Shorter than minimum length 5.", data["detail_message"].(map[string]interface{})["description"].(string))
				assert.Equal(t, "Shorter than minimum length 1.", data["detail_message"].(map[string]interface{})["price"].(string))
				assert.Equal(t, "Shorter than minimum length 1.", data["detail_message"].(map[string]interface{})["stock"].(string))
				assert.Equal(t, "Shorter than minimum length 1.", data["detail_message"].(map[string]interface{})["category_id"].(string))
			case "maximum":
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["name"].(string))
			case "danger file extension":
				assert.Equal(t, "Image must be between jpeg, png.", data["detail_message"].(map[string]interface{})["image"].(string))
			case "not valid file extension":
				assert.Equal(t, "Image must be between jpeg, png.", data["detail_message"].(map[string]interface{})["image"].(string))
			case "file cannot grater than 4 Mb":
				assert.Equal(t, "An image cannot greater than 4 Mb.", data["detail_message"].(map[string]interface{})["image"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestCreateProduct(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	category, _ := repo.categoriesRepo.GetCategoryByName(context.Background(), namee)

	tests := [...]struct {
		name       string
		payload    map[string]string
		expected   string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			payload:    map[string]string{"name": namee, "description": "asdasd", "image": "@/app/static/test_image/image.jpeg", "price": "1", "stock": "1", "category_id": "1"},
			expected:   "User not found.",
			token:      tokenNotFound,
			statusCode: 401,
		},
		{
			name:       "user not admin",
			payload:    map[string]string{"name": namee, "description": "asdasd", "image": "@/app/static/test_image/image.jpeg", "price": "1", "stock": "1", "category_id": "1"},
			expected:   "Only users with admin privileges can do this action.",
			token:      tokenGuest,
			statusCode: 401,
		},
		{
			name:       "category not found",
			payload:    map[string]string{"name": namee, "description": "asdasd", "image": "@/app/static/test_image/image.jpeg", "price": "1", "stock": "1", "category_id": "999999999"},
			expected:   "Category not found.",
			token:      tokenAdmin,
			statusCode: 404,
		},
		{
			name:       "success",
			payload:    map[string]string{"name": namee, "description": "asdasd", "image": "@/app/static/test_image/image.jpeg", "price": "1", "stock": "1", "category_id": strconv.Itoa(category.Id)},
			expected:   "Successfully add a new product.",
			token:      tokenAdmin,
			statusCode: 201,
		},
		{
			name:       "name duplicate",
			payload:    map[string]string{"name": namee, "description": "asdasd", "image": "@/app/static/test_image/image.jpeg", "price": "1", "stock": "1", "category_id": strconv.Itoa(category.Id)},
			expected:   "The name has already been taken.",
			token:      tokenAdmin,
			statusCode: 400,
		},
		{
			name:       "success 2",
			payload:    map[string]string{"name": namee_2, "description": "asdasd", "image": "@/app/static/test_image/image.jpeg", "price": "1", "stock": "1", "category_id": strconv.Itoa(category.Id)},
			expected:   "Successfully add a new product.",
			token:      tokenAdmin,
			statusCode: 201,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ct, b, err := createForm(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefixProduct, b)
			req.Header.Add("Authorization", "Bearer "+test.token)
			req.Header.Set("Content-Type", ct)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found", "user not admin":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_header"].(string))
			case "category not found", "success", "name duplicate":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestValidationUpdateProduct(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name    string
		payload map[string]string
	}{
		{
			name:    "minimum",
			payload: map[string]string{"name": "a", "description": "a", "image": "@/app/static/test_image/image.jpeg", "price": "-1", "stock": "-1", "category_id": "-1"},
		},
		{
			name:    "maximum",
			payload: map[string]string{"name": createMaximum(200), "description": createMaximum(200), "image": "@/app/static/test_image/image.jpeg"},
		},
		{
			name:    "danger file extension",
			payload: map[string]string{"image": "@/app/static/test_image/test.txt"},
		},
		{
			name:    "not valid file extension",
			payload: map[string]string{"image": "@/app/static/test_image/test.gif"},
		},
		{
			name:    "file cannot grater than 4 Mb",
			payload: map[string]string{"image": "@/app/static/test_image/size.png"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ct, b, err := createForm(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPut, prefixProduct+"/1", b)
			req.Header.Add("Authorization", "Bearer "+tokenAdmin)
			req.Header.Set("Content-Type", ct)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "minimum":
				assert.Equal(t, "Shorter than minimum length 3.", data["detail_message"].(map[string]interface{})["name"].(string))
				assert.Equal(t, "Shorter than minimum length 5.", data["detail_message"].(map[string]interface{})["description"].(string))
				assert.Equal(t, "Shorter than minimum length 1.", data["detail_message"].(map[string]interface{})["price"].(string))
				assert.Equal(t, "Shorter than minimum length 1.", data["detail_message"].(map[string]interface{})["stock"].(string))
				assert.Equal(t, "Shorter than minimum length 1.", data["detail_message"].(map[string]interface{})["category_id"].(string))
			case "maximum":
				assert.Equal(t, "Longer than maximum length 100.", data["detail_message"].(map[string]interface{})["name"].(string))
			case "danger file extension":
				assert.Equal(t, "Image must be between jpeg, png.", data["detail_message"].(map[string]interface{})["image"].(string))
			case "not valid file extension":
				assert.Equal(t, "Image must be between jpeg, png.", data["detail_message"].(map[string]interface{})["image"].(string))
			case "file cannot grater than 4 Mb":
				assert.Equal(t, "An image cannot greater than 4 Mb.", data["detail_message"].(map[string]interface{})["image"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}

}

func TestUpdateProduct(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	category, _ := repo.categoriesRepo.GetCategoryByName(context.Background(), namee)
	product, _ := repo.productsRepo.GetProductByName(context.Background(), namee)

	tests := [...]struct {
		name       string
		payload    map[string]string
		expected   string
		token      string
		url        string
		statusCode int
	}{
		{
			name:       "user not found",
			payload:    map[string]string{"name": namee, "description": "asdasd", "image": "@/app/static/test_image/image.jpeg", "price": "1", "stock": "1", "category_id": "1"},
			expected:   "User not found.",
			token:      tokenNotFound,
			url:        prefixProduct + "/1",
			statusCode: 401,
		},
		{
			name:       "user not admin",
			payload:    map[string]string{"name": namee, "description": "asdasd", "image": "@/app/static/test_image/image.jpeg", "price": "1", "stock": "1", "category_id": "1"},
			expected:   "Only users with admin privileges can do this action.",
			token:      tokenGuest,
			url:        prefixProduct + "/1",
			statusCode: 401,
		},
		{
			name:       "product not found",
			payload:    map[string]string{"name": namee, "description": "asdasd", "image": "@/app/static/test_image/image.jpeg", "price": "1", "stock": "1", "category_id": "1"},
			expected:   "Product not found.",
			token:      tokenAdmin,
			url:        prefixProduct + "/999999999",
			statusCode: 404,
		},
		{
			name:       "name already taken",
			payload:    map[string]string{"name": namee_2, "description": "asdasd", "image": "@/app/static/test_image/image.jpeg", "price": "1", "stock": "1", "category_id": "1"},
			expected:   "The name has already been taken.",
			token:      tokenAdmin,
			url:        prefixProduct + "/" + strconv.Itoa(product.Id),
			statusCode: 400,
		},
		{
			name:       "category not found",
			payload:    map[string]string{"name": namee, "description": "asdasd", "image": "@/app/static/test_image/image.jpeg", "price": "1", "stock": "1", "category_id": "999999999"},
			expected:   "Category not found.",
			token:      tokenAdmin,
			url:        prefixProduct + "/" + strconv.Itoa(product.Id),
			statusCode: 404,
		},
		{
			name:       "success",
			payload:    map[string]string{"name": namee, "description": "asdasd", "image": "@/app/static/test_image/image.jpeg", "price": "1", "stock": "1", "category_id": strconv.Itoa(category.Id)},
			expected:   "Successfully update the product.",
			token:      tokenAdmin,
			url:        prefixProduct + "/" + strconv.Itoa(product.Id),
			statusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ct, b, err := createForm(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPut, test.url, b)
			req.Header.Add("Authorization", "Bearer "+test.token)
			req.Header.Set("Content-Type", ct)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found", "user not admin":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_header"].(string))
			case "product not found", "name already taken", "category not found", "success":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestValidationGetAllProduct(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name string
		url  string
	}{
		{
			name: "empty",
			url:  prefixProduct,
		},
		{
			name: "required",
			url:  prefixProduct + "?page=&per_page=&q=&order_by=&category_id=",
		},
		{
			name: "type data",
			url:  prefixProduct + "?page=a&per_page=a&q=1&=order_by=a&category_id=a",
		},
		{
			name: "minimum",
			url:  prefixProduct + "?page=-1&per_page=-1&q=&order_by=&category_id=-1",
		},
		{
			name: "one of",
			url:  prefixProduct + "?page=1&per_page=1&order_by=a",
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
				assert.Equal(t, "Must be greater than or equal to 1.", data["detail_message"].(map[string]interface{})["category_id"].(string))
			case "one of":
				assert.Equal(t, "Must be one of: high_price, low_price.", data["detail_message"].(map[string]interface{})["order_by"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestGetAllProduct(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	req, _ := http.NewRequest(http.MethodGet, prefixProduct+"?page=1&per_page=1&q=t", nil)

	response := executeRequest(req, s)

	body, _ := io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	assert.NotNil(t, data["results"].(map[string]interface{})["data"])
	assert.Equal(t, 200, response.Result().StatusCode)
}

func TestGetProductBySlug(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	// get id
	product, _ := repo.productsRepo.GetProductByName(context.Background(), namee)

	tests := [...]struct {
		name string
		url  string
	}{
		{
			name: "not found",
			url:  prefixProduct + "/99999999",
		},
		{
			name: "success",
			url:  prefixProduct + "/" + product.Slug,
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
				assert.Equal(t, "Product not found.", data["detail_message"].(map[string]interface{})["_app"].(string))
				assert.Equal(t, 404, response.Result().StatusCode)
			case "success":
				assert.NotNil(t, data["results"])
				assert.Equal(t, 200, response.Result().StatusCode)
			}
		})
	}
}

func TestDeleteProduct(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	product, _ := repo.productsRepo.GetProductByName(context.Background(), namee)
	product_2, _ := repo.productsRepo.GetProductByName(context.Background(), namee_2)

	tests := [...]struct {
		name       string
		expected   string
		token      string
		url        string
		statusCode int
	}{
		{
			name:       "user not found",
			expected:   "User not found.",
			token:      tokenNotFound,
			url:        prefixProduct + "/1",
			statusCode: 401,
		},
		{
			name:       "user not admin",
			expected:   "Only users with admin privileges can do this action.",
			token:      tokenGuest,
			url:        prefixProduct + "/1",
			statusCode: 401,
		},
		{
			name:       "product not found",
			expected:   "Product not found.",
			token:      tokenAdmin,
			url:        prefixProduct + "/9999999999",
			statusCode: 404,
		},
		{
			name:       "success",
			expected:   "Successfully delete the product.",
			token:      tokenAdmin,
			url:        prefixProduct + "/" + strconv.Itoa(product.Id),
			statusCode: 200,
		},
		{
			name:       "success",
			expected:   "Successfully delete the product.",
			token:      tokenAdmin,
			url:        prefixProduct + "/" + strconv.Itoa(product_2.Id),
			statusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodDelete, test.url, nil)
			req.Header.Add("Authorization", "Bearer "+test.token)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found", "user not admin":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_header"].(string))
			case "product not found", "success":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)

		})
	}
}

func TestDownProduct(t *testing.T) {
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

	product, _ := repo.productsRepo.GetProductByName(context.Background(), namee)
	repo.productsRepo.Delete(context.Background(), product.Id)

	product, _ = repo.productsRepo.GetProductByName(context.Background(), namee_2)
	repo.productsRepo.Delete(context.Background(), product.Id)
}
