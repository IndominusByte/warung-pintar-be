package tests

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	ordersentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/orders"
	"github.com/stretchr/testify/assert"
)

const (
	prefixOrder = "/orders"
)

func TestValidationCreateOrder(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name    string
		payload map[string]string
	}{
		{
			name:    "required file",
			payload: map[string]string{},
		},
		{
			name:    "required form",
			payload: map[string]string{"proof_of_payment": "@/app/static/test_image/image.jpeg"},
		},
		{
			name:    "minimum",
			payload: map[string]string{"proof_of_payment": "@/app/static/test_image/image.jpeg", "fullname": "a", "address": "a"},
		},
		{
			name:    "maximum",
			payload: map[string]string{"proof_of_payment": "@/app/static/test_image/image.jpeg", "fullname": createMaximum(200)},
		},
		{
			name:    "invalid phone",
			payload: map[string]string{"proof_of_payment": "@/app/static/test_image/image.jpeg", "phone": "878622222"},
		},
		{
			name:    "danger file extension",
			payload: map[string]string{"proof_of_payment": "@/app/static/test_image/test.txt"},
		},
		{
			name:    "not valid file extension",
			payload: map[string]string{"proof_of_payment": "@/app/static/test_image/test.gif"},
		},
		{
			name:    "file cannot grater than 4 Mb",
			payload: map[string]string{"proof_of_payment": "@/app/static/test_image/size.png"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ct, b, err := createForm(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefixOrder, b)
			req.Header.Add("Authorization", "Bearer "+tokenGuest)
			req.Header.Set("Content-Type", ct)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "required file":
				assert.Equal(t, "Image is required.", data["detail_message"].(map[string]interface{})["proof_of_payment"].(string))
			case "required form":
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
			case "danger file extension":
				assert.Equal(t, "Image must be between jpeg, png.", data["detail_message"].(map[string]interface{})["proof_of_payment"].(string))
			case "not valid file extension":
				assert.Equal(t, "Image must be between jpeg, png.", data["detail_message"].(map[string]interface{})["proof_of_payment"].(string))
			case "file cannot grater than 4 Mb":
				assert.Equal(t, "An image cannot greater than 4 Mb.", data["detail_message"].(map[string]interface{})["proof_of_payment"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestCreateOrder(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	// update qty cart
	user, _ := repo.authRepo.GetUserByEmail(context.Background(), email_2)
	cart, _ := repo.cartsRepo.GetCartByUserIdAndProductId(context.Background(), user.Id, productId)
	cart.Qty = 99
	repo.cartsRepo.Update(context.Background(), cart)

	tests := [...]struct {
		name       string
		payload    map[string]string
		expected   string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			payload:    map[string]string{"proof_of_payment": "@/app/static/test_image/image.jpeg", "fullname": "asdasd", "phone": "08786226533", "address": "asdasd"},
			expected:   "User not found.",
			token:      tokenNotFound,
			statusCode: 401,
		},
		{
			name:       "item not found",
			payload:    map[string]string{"proof_of_payment": "@/app/static/test_image/image.jpeg", "fullname": "asdasd", "phone": "08786226533", "address": "asdasd"},
			expected:   "Ups, item in payment not found.",
			token:      tokenAdmin,
			statusCode: 404,
		},
		{
			name:       "qty exceed",
			payload:    map[string]string{"proof_of_payment": "@/app/static/test_image/image.jpeg", "fullname": "asdasd", "phone": "08786226533", "address": "asdasd"},
			expected:   fmt.Sprintf("Available stock: %d, please reduce quantity product '%s'", 1, namee),
			token:      tokenGuest,
			statusCode: 400,
		},
		{
			name:       "success",
			payload:    map[string]string{"proof_of_payment": "@/app/static/test_image/image.jpeg", "fullname": "asdasd", "phone": "08786226533", "address": "asdasd"},
			expected:   "Successfully save the order.",
			token:      tokenGuest,
			statusCode: 201,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ct, b, err := createForm(test.payload)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPost, prefixOrder, b)
			req.Header.Add("Authorization", "Bearer "+test.token)
			req.Header.Set("Content-Type", ct)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_header"].(string))
			case "qty exceed":
				cart.Qty = 1
				repo.cartsRepo.Update(context.Background(), cart)
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			default:
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestValidationGetOrderAdmin(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name string
		url  string
	}{
		{
			name: "empty",
			url:  prefixOrder,
		},
		{
			name: "required",
			url:  prefixOrder + "?page=&per_page=&status=",
		},
		{
			name: "type data",
			url:  prefixOrder + "?page=a&per_page=a&status=1",
		},
		{
			name: "minimum",
			url:  prefixOrder + "?page=-1&per_page=-1&status=",
		},
		{
			name: "one of",
			url:  prefixOrder + "?page=1&per_page=1&status=a",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodGet, test.url, nil)
			req.Header.Add("Authorization", "Bearer "+tokenAdmin)

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
			case "one of":
				assert.Equal(t, "Must be one of: 'ongoing', 'reject', 'on the way', 'success'.", data["detail_message"].(map[string]interface{})["status"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestGetOrderAdmin(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	req, _ := http.NewRequest(http.MethodGet, prefixOrder+"?page=1&per_page=1", nil)
	req.Header.Add("Authorization", "Bearer "+tokenAdmin)

	response := executeRequest(req, s)

	body, _ := io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	assert.NotNil(t, data["results"].(map[string]interface{})["data"])
	assert.Equal(t, 200, response.Result().StatusCode)
}

func TestValidationGetOrderGuest(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	tests := [...]struct {
		name string
		url  string
	}{
		{
			name: "empty",
			url:  prefixOrder + "/mine",
		},
		{
			name: "required",
			url:  prefixOrder + "/mine" + "?page=&per_page=&status=",
		},
		{
			name: "type data",
			url:  prefixOrder + "/mine" + "?page=a&per_page=a&status=1",
		},
		{
			name: "minimum",
			url:  prefixOrder + "/mine" + "?page=-1&per_page=-1&status=",
		},
		{
			name: "one of",
			url:  prefixOrder + "/mine" + "?page=1&per_page=1&status=a",
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
			case "empty", "required":
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["page"].(string))
				assert.Equal(t, "Missing data for required field.", data["detail_message"].(map[string]interface{})["per_page"].(string))
			case "type data":
				assert.Equal(t, "Invalid input type.", data["detail_message"].(map[string]interface{})["_body"].(string))
			case "minimum":
				assert.Equal(t, "Must be greater than or equal to 1.", data["detail_message"].(map[string]interface{})["page"].(string))
				assert.Equal(t, "Must be greater than or equal to 1.", data["detail_message"].(map[string]interface{})["per_page"].(string))
			case "one of":
				assert.Equal(t, "Must be one of: 'ongoing', 'reject', 'on the way', 'success'.", data["detail_message"].(map[string]interface{})["status"].(string))
			}
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestGetOrderGuest(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

	req, _ := http.NewRequest(http.MethodGet, prefixOrder+"/mine"+"?page=1&per_page=1", nil)
	req.Header.Add("Authorization", "Bearer "+tokenGuest)

	response := executeRequest(req, s)

	body, _ := io.ReadAll(response.Result().Body)
	json.Unmarshal(body, &data)

	assert.NotNil(t, data["results"].(map[string]interface{})["data"])
	assert.Equal(t, 200, response.Result().StatusCode)
}

func TestSetRejectOrder(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	user, _ := repo.authRepo.GetUserByEmail(context.Background(), email_2)
	order, _ := repo.ordersRepo.GetOrderByUserIdLimit(context.Background(), user.Id)

	repo.ordersRepo.UpdateOrder(context.Background(), &ordersentity.Order{
		Id:     order.Id,
		Status: "success",
	})

	tests := [...]struct {
		name       string
		url        string
		expected   string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			url:        prefixOrder + fmt.Sprintf("/set-reject/%d", order.Id),
			expected:   "User not found.",
			token:      tokenNotFound,
			statusCode: 401,
		},
		{
			name:       "user not admin",
			url:        prefixOrder + fmt.Sprintf("/set-reject/%d", order.Id),
			expected:   "Only users with admin privileges can do this action.",
			token:      tokenGuest,
			statusCode: 401,
		},
		{
			name:       "order not found",
			url:        prefixOrder + fmt.Sprintf("/set-reject/%d", 999999),
			expected:   "Order not found.",
			token:      tokenAdmin,
			statusCode: 404,
		},
		{
			name:       "status not ongoing",
			url:        prefixOrder + fmt.Sprintf("/set-reject/%d", order.Id),
			expected:   "Cannot change status rejected if status other than ongoing.",
			token:      tokenAdmin,
			statusCode: 400,
		},
		{
			name:       "success",
			url:        prefixOrder + fmt.Sprintf("/set-reject/%d", order.Id),
			expected:   "Successfully set the order to reject.",
			token:      tokenAdmin,
			statusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPut, test.url, nil)
			req.Header.Add("Authorization", "Bearer "+test.token)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found", "user not admin":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_header"].(string))
			case "status not ongoing":
				repo.ordersRepo.UpdateOrder(context.Background(), &ordersentity.Order{
					Id:     order.Id,
					Status: "ongoing",
				})
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			default:
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestValidationSetOnTheWayOrder(t *testing.T) {
	_, s := setupEnvironment()

	var data map[string]interface{}

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
			form:     map[string]string{"no_receipt": "@/app/static/test_image/test.txt"},
		},
		{
			name:     "not valid file extension",
			expected: "Image must be between jpeg, png.",
			form:     map[string]string{"no_receipt": "@/app/static/test_image/test.gif"},
		},
		{
			name:     "file cannot grater than 4 Mb",
			expected: "An image cannot greater than 4 Mb.",
			form:     map[string]string{"no_receipt": "@/app/static/test_image/size.png"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ct, b, err := createForm(test.form)
			if err != nil {
				panic(err)
			}

			req, _ := http.NewRequest(http.MethodPut, prefixOrder+"/set-on-the-way/999999", b)
			req.Header.Add("Authorization", "Bearer "+tokenAdmin)
			req.Header.Set("Content-Type", ct)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["no_receipt"].(string))
			assert.Equal(t, 422, response.Result().StatusCode)
		})
	}
}

func TestSetOnTheWayOrder(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	user, _ := repo.authRepo.GetUserByEmail(context.Background(), email_2)
	order, _ := repo.ordersRepo.GetOrderByUserIdLimit(context.Background(), user.Id)

	repo.ordersRepo.UpdateOrder(context.Background(), &ordersentity.Order{
		Id:     order.Id,
		Status: "success",
	})

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
			url:        prefixOrder + fmt.Sprintf("/set-on-the-way/%d", order.Id),
			payload:    map[string]string{"no_receipt": "@/app/static/test_image/image.jpeg"},
			expected:   "User not found.",
			token:      tokenNotFound,
			statusCode: 401,
		},
		{
			name:       "user not admin",
			url:        prefixOrder + fmt.Sprintf("/set-on-the-way/%d", order.Id),
			payload:    map[string]string{"no_receipt": "@/app/static/test_image/image.jpeg"},
			expected:   "Only users with admin privileges can do this action.",
			token:      tokenGuest,
			statusCode: 401,
		},
		{
			name:       "order not found",
			url:        prefixOrder + fmt.Sprintf("/set-on-the-way/%d", 999999),
			payload:    map[string]string{"no_receipt": "@/app/static/test_image/image.jpeg"},
			expected:   "Order not found.",
			token:      tokenAdmin,
			statusCode: 404,
		},
		{
			name:       "status not ongoing",
			url:        prefixOrder + fmt.Sprintf("/set-on-the-way/%d", order.Id),
			payload:    map[string]string{"no_receipt": "@/app/static/test_image/image.jpeg"},
			expected:   "Cannot change status on the way if status other than ongoing.",
			token:      tokenAdmin,
			statusCode: 400,
		},
		{
			name:       "success",
			url:        prefixOrder + fmt.Sprintf("/set-on-the-way/%d", order.Id),
			payload:    map[string]string{"no_receipt": "@/app/static/test_image/image.jpeg"},
			expected:   "Successfully set the order to on the way.",
			token:      tokenAdmin,
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
			case "status not ongoing":
				repo.ordersRepo.UpdateOrder(context.Background(), &ordersentity.Order{
					Id:     order.Id,
					Status: "ongoing",
				})
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			default:
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestSetSuccessOrder(t *testing.T) {
	repo, s := setupEnvironment()

	var data map[string]interface{}

	user, _ := repo.authRepo.GetUserByEmail(context.Background(), email_2)
	order, _ := repo.ordersRepo.GetOrderByUserIdLimit(context.Background(), user.Id)

	repo.ordersRepo.UpdateOrder(context.Background(), &ordersentity.Order{
		Id:     order.Id,
		Status: "success",
	})

	tests := [...]struct {
		name       string
		url        string
		expected   string
		token      string
		statusCode int
	}{
		{
			name:       "user not found",
			url:        prefixOrder + fmt.Sprintf("/set-success/%d", order.Id),
			expected:   "User not found.",
			token:      tokenNotFound,
			statusCode: 401,
		},
		{
			name:       "order not found",
			url:        prefixOrder + fmt.Sprintf("/set-success/%d", 999999),
			expected:   "Order not found.",
			token:      tokenGuest,
			statusCode: 404,
		},
		{
			name:       "status not on the way",
			url:        prefixOrder + fmt.Sprintf("/set-success/%d", order.Id),
			expected:   "Cannot change status success if status other than on the way.",
			token:      tokenGuest,
			statusCode: 400,
		},
		{
			name:       "user not same as order",
			url:        prefixOrder + fmt.Sprintf("/set-success/%d", order.Id),
			expected:   "User doesn't have this order.",
			token:      tokenAdmin,
			statusCode: 400,
		},
		{
			name:       "success",
			url:        prefixOrder + fmt.Sprintf("/set-success/%d", order.Id),
			expected:   "Successfully set the order to success.",
			token:      tokenGuest,
			statusCode: 200,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, _ := http.NewRequest(http.MethodPut, test.url, nil)
			req.Header.Add("Authorization", "Bearer "+test.token)

			response := executeRequest(req, s)

			body, _ := io.ReadAll(response.Result().Body)
			json.Unmarshal(body, &data)

			switch test.name {
			case "user not found", "user not admin":
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_header"].(string))
			case "status not on the way":
				repo.ordersRepo.UpdateOrder(context.Background(), &ordersentity.Order{
					Id:     order.Id,
					Status: "on the way",
				})
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			default:
				assert.Equal(t, test.expected, data["detail_message"].(map[string]interface{})["_app"].(string))
			}
			assert.Equal(t, test.statusCode, response.Result().StatusCode)
		})
	}
}

func TestDownOrder(t *testing.T) {
	repo, _ := setupEnvironment()

	user, _ := repo.authRepo.GetUserByEmail(context.Background(), email)
	userConfirm, _ := repo.authRepo.GetUserConfirmByUserId(context.Background(), user.Id)

	repo.authRepo.DeleteUserConfirm(context.Background(), userConfirm.Id)
	repo.authRepo.DeleteUser(context.Background(), user.Id)

	user, _ = repo.authRepo.GetUserByEmail(context.Background(), email_2)
	userConfirm, _ = repo.authRepo.GetUserConfirmByUserId(context.Background(), user.Id)

	repo.cartsRepo.DeleteByUserId(context.Background(), user.Id)

	repo.authRepo.DeleteUserConfirm(context.Background(), userConfirm.Id)
	repo.authRepo.DeleteUser(context.Background(), user.Id)

	category, _ := repo.categoriesRepo.GetCategoryByName(context.Background(), namee)
	repo.categoriesRepo.Delete(context.Background(), category.Id)

	category, _ = repo.categoriesRepo.GetCategoryByName(context.Background(), namee_2)
	repo.categoriesRepo.Delete(context.Background(), category.Id)

	repo.productsRepo.Delete(context.Background(), productId)
	repo.productsRepo.Delete(context.Background(), productId2)
}
