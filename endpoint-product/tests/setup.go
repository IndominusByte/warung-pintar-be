package tests

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"

	"github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/config"
	handler_http "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/endpoint/http/handler"
	authrepo "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/repo/auth"
	categoriesrepo "github.com/IndominusByte/warung-pintar-be/endpoint-product/internal/repo/categories"
)

type setupRepo struct {
	authRepo       authrepo.RepoAuth
	categoriesRepo categoriesrepo.RepoCategories
}

func setupEnvironment() (*setupRepo, *handler_http.Server) {
	// init config
	cfg, err := config.New()
	if err != nil {
		panic(err)
	}
	// connect the db
	db, err := config.DBConnect(cfg)
	if err != nil {
		panic(err)
	}
	// connect redis
	redisCli, err := config.RedisConnect(cfg)
	if err != nil {
		panic(err)
	}
	// mount router
	r := handler_http.CreateNewServer(db, redisCli, cfg)
	if err := r.MountHandlers(); err != nil {
		panic(err)
	}
	// you can insert your behaviors here
	authRepo, _ := authrepo.New(db)
	categoriesRepo, _ := categoriesrepo.New(db)

	setuprepo := setupRepo{
		authRepo:       *authRepo,
		categoriesRepo: *categoriesRepo,
	}

	return &setuprepo, r
}

// executeRequest, creates a new ResponseRecorder
// then executes the request by calling ServeHTTP in the router
// after which the handler writes the response to the response recorder
// which we can then inspect.
func executeRequest(req *http.Request, s *handler_http.Server) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	s.Router.ServeHTTP(rr, req)

	return rr
}

func createForm(form map[string]string) (string, io.Reader, error) {
	body := new(bytes.Buffer)
	mp := multipart.NewWriter(body)
	defer mp.Close()
	for key, val := range form {
		if strings.HasPrefix(val, "@") {
			val = val[1:]
			if len(val) < 1 {
				mp.CreateFormFile(key, "")
				continue
			}
			file, err := os.Open(val)
			if err != nil {
				return "", nil, err
			}
			defer file.Close()
			filename := strings.Split(val, "/")
			part, err := mp.CreateFormFile(key, filename[len(filename)-1])
			if err != nil {
				return "", nil, err
			}
			io.Copy(part, file)
		} else {
			mp.WriteField(key, val)
		}
	}
	return mp.FormDataContentType(), body, nil
}

func createMaximum(length int) string {
	word := ""
	for i := 0; i < length; i++ {
		word += "a"
	}
	return word
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}
