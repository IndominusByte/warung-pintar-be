package products

import (
	"context"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/creent-production/cdk-go/magicimage"
	"github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/constant"
	productsentity "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/entity/products"
	"github.com/creent-production/cdk-go/response"
	"github.com/creent-production/cdk-go/validation"
	"github.com/go-chi/jwtauth"
	"github.com/gosimple/slug"
)

type ProductsUsecase struct {
	productsRepo   productsRepo
	categoriesRepo categoriesRepo
	authRepo       authRepo
}

func NewProductsUsecase(productRepo productsRepo, categoryRepo categoriesRepo, authRepo authRepo) *ProductsUsecase {
	return &ProductsUsecase{
		productsRepo:   productRepo,
		categoriesRepo: categoryRepo,
		authRepo:       authRepo,
	}
}

func (uc *ProductsUsecase) Create(ctx context.Context,
	rw http.ResponseWriter, file *multipart.Form, payload *productsentity.FormCreateUpdateSchema) {

	magicImage := magicimage.New(file)
	if err := magicImage.ValidateSingleImage("image"); err != nil {
		response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
			"image": err.Error(),
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

	if user.Role != "admin" {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: fmt.Sprintf(constant.PrivilegesOnly, "admin"),
		})
		return
	}

	payload.Slug = slug.Make(payload.Name)

	// check name duplicate
	if _, err := uc.productsRepo.GetProductBySlug(ctx, payload.Slug); err == nil {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: fmt.Sprintf(constant.AlreadyTaken, "name"),
		})
		return
	}
	// check category id not found
	if _, err := uc.categoriesRepo.GetCategoryById(ctx, payload.CategoryId); err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Category not found.",
		})
		return
	}

	magicImage.SaveImages(500, 500, "/app/static/products", true)
	payload.Image = magicImage.FileNames[0]

	// insert into db
	uc.productsRepo.Insert(ctx, payload)

	response.WriteJSONResponse(rw, 201, nil, map[string]interface{}{
		constant.App: "Successfully add a new product.",
	})
}

func (uc *ProductsUsecase) Update(ctx context.Context, rw http.ResponseWriter,
	file *multipart.Form, payload *productsentity.FormCreateUpdateSchema, productId int) {

	magicImage := magicimage.New(file)
	magicImage.Required = false
	if err := magicImage.ValidateSingleImage("image"); err != nil {
		response.WriteJSONResponse(rw, 422, nil, map[string]interface{}{
			"image": err.Error(),
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

	if user.Role != "admin" {
		response.WriteJSONResponse(rw, 401, nil, map[string]interface{}{
			constant.Header: fmt.Sprintf(constant.PrivilegesOnly, "admin"),
		})
		return
	}

	product, err := uc.productsRepo.GetProductById(ctx, productId)
	if err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Product not found.",
		})
		return
	}

	payload.Slug = slug.Make(payload.Name)

	// check name duplicate
	if _, err := uc.productsRepo.GetProductBySlug(ctx, payload.Slug); err == nil && product.Slug != payload.Slug {
		response.WriteJSONResponse(rw, 400, nil, map[string]interface{}{
			constant.App: fmt.Sprintf(constant.AlreadyTaken, "name"),
		})
		return
	}
	// check category id not found
	if _, err := uc.categoriesRepo.GetCategoryById(ctx, payload.CategoryId); err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Category not found.",
		})
		return
	}

	if _, ok := file.File["image"]; ok {
		if len(product.Image) > 0 {
			magicimage.DeleteFolderAndFile(fmt.Sprintf("/app/static/products/%s", product.Image))
		}
		magicImage.SaveImages(500, 500, "/app/static/products", true)
		payload.Image = magicImage.FileNames[0]
	}

	// update in db
	payload.Id = product.Id
	uc.productsRepo.Update(ctx, payload)

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "Successfully update the product.",
	})
}

func (uc *ProductsUsecase) Delete(ctx context.Context, rw http.ResponseWriter, productId int) {
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

	product, err := uc.productsRepo.GetProductById(ctx, productId)
	if err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Product not found.",
		})
		return
	}

	// delete guardian
	uc.productsRepo.Delete(ctx, product.Id)
	if len(product.Image) > 0 {
		magicimage.DeleteFolderAndFile(fmt.Sprintf("/app/static/products/%s", product.Image))
	}

	response.WriteJSONResponse(rw, 200, nil, map[string]interface{}{
		constant.App: "Successfully delete the product.",
	})
}

func (uc *ProductsUsecase) GetAll(ctx context.Context, rw http.ResponseWriter, payload *productsentity.QueryParamAllProductSchema) {
	if err := validation.StructValidate(payload); err != nil {
		response.WriteJSONResponse(rw, 422, nil, err)
		return
	}

	results, _ := uc.productsRepo.GetAllProductPaginate(ctx, payload)

	response.WriteJSONResponse(rw, 200, results, nil)
}

func (uc *ProductsUsecase) GetBySlug(ctx context.Context, rw http.ResponseWriter, slug string) {
	product, err := uc.productsRepo.GetProductBySlug(ctx, slug)
	if err != nil {
		response.WriteJSONResponse(rw, 404, nil, map[string]interface{}{
			constant.App: "Product not found.",
		})
		return
	}

	response.WriteJSONResponse(rw, 200, product, nil)
}
