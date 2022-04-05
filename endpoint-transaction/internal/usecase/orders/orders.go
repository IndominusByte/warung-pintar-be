package orders

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/constant"
	"github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/carts"
	ordersentity "github.com/IndominusByte/warung-pintar-be/endpoint-transaction/internal/entity/orders"
	"github.com/creent-production/cdk-go/magicimage"
	"github.com/creent-production/cdk-go/parser"
	"github.com/creent-production/cdk-go/response"
	"github.com/creent-production/cdk-go/validation"
	"github.com/go-chi/jwtauth"
	"github.com/gomodule/redigo/redis"
	"gopkg.in/guregu/null.v4"
)

type OrdersUsecase struct {
	ordersRepo ordersRepo
	authRepo   authRepo
	cartsRepo  cartsRepo
}

func NewOrdersUsecase(orderRepo ordersRepo, authRepo authRepo, cartRepo cartsRepo) *OrdersUsecase {
	return &OrdersUsecase{
		ordersRepo: orderRepo,
		authRepo:   authRepo,
		cartsRepo:  cartRepo,
	}
}

func (uc *OrdersUsecase) Create(ctx context.Context, rw http.ResponseWriter,
	redisCli *redis.Pool, file *multipart.Form, payload *ordersentity.FormCreateSchema) {

	magicImage := magicimage.New(file)
	if err := magicImage.ValidateSingleImage("proof_of_payment"); err != nil {
		response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
			"proof_of_payment": err.Error(),
		})
		return
	}

	if err := validation.StructValidate(payload); err != nil {
		response.WriteJSONResponse(rw, 422, nil, err)
		return
	}

	_, claims, _ := jwtauth.FromContext(ctx)
	sub, _ := strconv.Atoi(claims["sub"].(string))

	user, err := uc.authRepo.GetUserById(ctx, sub)
	if err != nil {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: constant.UserNotFound,
		})
		return
	}

	productData, _ := uc.cartsRepo.ItemInPayment(ctx, redisCli, user.Id)

	if len(productData) < 1 {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Ups, item in payment not found.",
		})
		return
	}

	payload.TotalAmount = 0
	payload.UserId = user.Id
	for _, product := range productData {
		// check qty exceed product stock
		if product.CartQty > product.ProductStock {
			response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
				constant.App: fmt.Sprintf("Available stock: %d, please reduce quantity product '%s'", product.ProductStock, product.ProductName),
			})
			return
		}
		payload.TotalAmount += product.CartQty * product.ProductPrice
	}

	magicImage.SaveImages(500, 500, "/app/static/proof_payments", false)
	payload.ProofOfPayment = magicImage.FileNames[0]

	// insert into db
	orderId := uc.ordersRepo.Insert(ctx, payload)

	for _, orderItem := range productData {
		uc.ordersRepo.InsertItem(ctx, &ordersentity.OrderItem{
			Notes:     orderItem.CartNotes,
			Qty:       orderItem.CartQty,
			Price:     orderItem.ProductPrice,
			ProductId: orderItem.CartProductId,
			OrderId:   orderId,
		})

	}

	// delete cart db and redis payment
	conn := redisCli.Get()
	defer conn.Close()

	item, _ := redis.String(conn.Do("GET", fmt.Sprintf("checkout:%d", user.Id)))
	listid, _ := parser.ParseSliceStrToSliceInt(strings.Split(item, ","))
	uc.cartsRepo.Delete(ctx, &carts.JsonMultipleSchema{
		UserId: user.Id,
		ListId: listid,
	})

	conn.Do("DEL", fmt.Sprintf("checkout:%d", payload.UserId))

	response.WriteJSONResponse(rw, 201, nil, map[string]interface{}{
		constant.App: "Successfully save the order.",
	})
}

func (uc *OrdersUsecase) SetReject(ctx context.Context, rw http.ResponseWriter, orderId int) {
	_, claims, _ := jwtauth.FromContext(ctx)
	sub, _ := strconv.Atoi(claims["sub"].(string))

	user, err := uc.authRepo.GetUserById(ctx, sub)
	if err != nil {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: constant.UserNotFound,
		})
		return
	}

	if user.Role != "admin" {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: fmt.Sprintf(constant.PrivilegesOnly, "admin"),
		})
		return
	}

	order, err := uc.ordersRepo.GetOrderById(ctx, orderId)
	if err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Order not found.",
		})
		return
	}

	if order.Status != "ongoing" {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: "Cannot change status rejected if status other than ongoing.",
		})
		return
	}

	// update status
	uc.ordersRepo.UpdateOrder(ctx, &ordersentity.Order{
		Id:     order.Id,
		Status: "reject",
	})

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "Successfully set the order to reject.",
	})
}

func (uc *OrdersUsecase) SetOnTheWay(ctx context.Context, rw http.ResponseWriter, orderId int, file *multipart.Form) {
	magicImage := magicimage.New(file)
	if err := magicImage.ValidateSingleImage("no_receipt"); err != nil {
		response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
			"no_receipt": err.Error(),
		})
		return
	}
	_, claims, _ := jwtauth.FromContext(ctx)
	sub, _ := strconv.Atoi(claims["sub"].(string))

	user, err := uc.authRepo.GetUserById(ctx, sub)
	if err != nil {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: constant.UserNotFound,
		})
		return
	}

	if user.Role != "admin" {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: fmt.Sprintf(constant.PrivilegesOnly, "admin"),
		})
		return
	}

	order, err := uc.ordersRepo.GetOrderById(ctx, orderId)
	if err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Order not found.",
		})
		return
	}

	if order.Status != "ongoing" {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: "Cannot change status on the way if status other than ongoing.",
		})
		return
	}

	// update status
	magicImage.SaveImages(500, 500, "/app/static/no_receipts", false)
	uc.ordersRepo.UpdateOrder(ctx, &ordersentity.Order{
		Id:        order.Id,
		Status:    "on the way",
		NoReceipt: null.StringFrom(magicImage.FileNames[0]),
	})

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "Successfully set the order to on the way.",
	})
}

func (uc *OrdersUsecase) SetSuccess(ctx context.Context, rw http.ResponseWriter, orderId int) {
	_, claims, _ := jwtauth.FromContext(ctx)
	sub, _ := strconv.Atoi(claims["sub"].(string))

	user, err := uc.authRepo.GetUserById(ctx, sub)
	if err != nil {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: constant.UserNotFound,
		})
		return
	}

	order, err := uc.ordersRepo.GetOrderById(ctx, orderId)
	if err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Order not found.",
		})
		return
	}

	if order.Status != "on the way" {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: "Cannot change status success if status other than on the way.",
		})
		return
	}

	if user.Id != order.UserId {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: "User doesn't have this order.",
		})
		return
	}

	// update status
	uc.ordersRepo.UpdateOrder(ctx, &ordersentity.Order{
		Id:     order.Id,
		Status: "success",
	})

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "Successfully set the order to success.",
	})
}
