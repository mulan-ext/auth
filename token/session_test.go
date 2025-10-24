package token_test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mulan-ext/auth/token"
)

func handlerInfo(c *gin.Context) {
	v := token.Default(c).Data()
	buf, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		c.String(500, "Marshal"+err.Error())
		return
	}
	err = json.Unmarshal(buf, v)
	if err != nil {
		c.String(500, "Unmarshal"+err.Error())
		return
	}
	c.JSON(200, v)
}

func handlerLogin(c *gin.Context) {
	v := token.Default(c)
	v.SetID(1)
	v.SetAccount("test")
	v.SetRoles([]string{"admin"})
	v.SetValues("aaaa", "aaaaa")
	v.SetValues("vvvv", "asdveasd")
	v.Save()
	c.String(200, v.Token())
}

func setupRouter(store token.Store) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(token.Init("token", store))
	r.GET("/info", handlerInfo)
	r.GET("/login", handlerLogin)
	return r
}

func TestTokenMem(t *testing.T) {
	store := token.NewMemStore()
	r := setupRouter(store)

	// 构建返回值
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/login", nil)
	r.ServeHTTP(w1, req1)
	rsp1 := w1.Result()
	body, _ := io.ReadAll(rsp1.Body)
	tokenStr := string(body)
	t.Log(tokenStr)

	req2, _ := http.NewRequest("GET", "/info", nil)
	req2.Header.Set("Authorization", "Bearer "+tokenStr)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	rsp2 := w2.Result()
	body, _ = io.ReadAll(rsp2.Body)
	t.Log(string(body))
}

func TestTokenFs(t *testing.T) {
	store, err := token.NewFsStore("/tmp/fs")
	if err != nil {
		t.Fatal(err)
	}
	r := setupRouter(store)

	// 构建返回值
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest("GET", "/login", nil)
	r.ServeHTTP(w1, req1)
	rsp1 := w1.Result()
	body, _ := io.ReadAll(rsp1.Body)
	tokenStr := string(body)
	t.Log(tokenStr)

	req2, _ := http.NewRequest("GET", "/info", nil)
	req2.Header.Set("Authorization", "Bearer "+tokenStr)
	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, req2)
	rsp2 := w2.Result()
	body, _ = io.ReadAll(rsp2.Body)
	t.Log(string(body))
}
