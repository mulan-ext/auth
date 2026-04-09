package session_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/mulan-ext/auth/session"
)

func buildInitNameRouter(t *testing.T, cfg *session.Config) *gin.Engine {
	t.Helper()

	gin.SetMode(gin.TestMode)

	mw, err := session.Init(cfg)
	if err != nil {
		t.Fatalf("init session middleware failed: %v", err)
	}

	r := gin.New()
	r.Use(mw)

	r.GET("/login", func(c *gin.Context) {
		sess := session.Default(c)
		sess.SetID(1001)
		sess.SetAccount("tester")
		sess.SetRoles([]string{"admin"})
		if err := sess.Save(); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		c.String(http.StatusOK, sess.Token())
	})

	r.GET("/me", session.AuthMW(), func(c *gin.Context) {
		c.String(http.StatusOK, c.GetString(session.CtxKeyAccount))
	})

	return r
}

func TestInitUsesConfiguredNameForSetCookie(t *testing.T) {
	const sessionName = "auth_token"

	r := buildInitNameRouter(t, &session.Config{
		Name:   sessionName,
		Driver: "memory",
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/login", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("unexpected status: got %d want %d", w.Code, http.StatusOK)
	}

	setCookie := w.Header().Get("Set-Cookie")
	if setCookie == "" {
		t.Fatal("expected Set-Cookie header")
	}
	if !strings.Contains(setCookie, sessionName+"=") {
		t.Fatalf("expected Set-Cookie to use configured name %q, got %q", sessionName, setCookie)
	}
}

func TestInitUsesConfiguredNameForHeaderExtraction(t *testing.T) {
	const sessionName = "auth_token"

	r := buildInitNameRouter(t, &session.Config{
		Name:   sessionName,
		Driver: "memory",
	})

	loginW := httptest.NewRecorder()
	loginReq, _ := http.NewRequest(http.MethodGet, "/login", nil)
	r.ServeHTTP(loginW, loginReq)

	if loginW.Code != http.StatusOK {
		t.Fatalf("unexpected login status: got %d want %d", loginW.Code, http.StatusOK)
	}

	token := strings.TrimSpace(loginW.Body.String())
	if token == "" {
		t.Fatal("expected token from login response")
	}

	meW := httptest.NewRecorder()
	meReq, _ := http.NewRequest(http.MethodGet, "/me", nil)
	meReq.Header.Set(sessionName, token)
	r.ServeHTTP(meW, meReq)

	if meW.Code != http.StatusOK {
		t.Fatalf("expected configured header name to authenticate request, got status %d", meW.Code)
	}

	if body := strings.TrimSpace(meW.Body.String()); body != "tester" {
		t.Fatalf("unexpected response body: got %q want %q", body, "tester")
	}
}
