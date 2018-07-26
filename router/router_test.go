package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/urfave/negroni"
)

func TestHandler(t *testing.T) {
	router := New()

	h := func(rw http.ResponseWriter, _ *http.Request) {
		rw.WriteHeader(http.StatusTeapot)
	}

	router.Handler("GET", "/", Wrap(h))
	router.Handler("GET", "/hello", Wrap(h))
	router.Handler("GET", "/hello/world", Wrap(h))
	router.Handler("GET", "/hello/world/elvizlai", Wrap(h))

	r := httptest.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, r)
	if rw.Code != http.StatusTeapot {
		t.Error("Test Handle failed")
	}

	r = httptest.NewRequest("GET", "//hello", nil)
	rw = httptest.NewRecorder()
	router.ServeHTTP(rw, r)
	if rw.Code != http.StatusMovedPermanently {
		t.Error("Test Handle failed", rw.Code)
	}
	if rw.Header().Get("Location") != "/hello" {
		t.Error("Test Handle failed", rw.Header())
	}

	r = httptest.NewRequest("GET", "/hello//world", nil)
	rw = httptest.NewRecorder()
	router.ServeHTTP(rw, r)
	if rw.Code != http.StatusMovedPermanently {
		t.Error("Test Handle failed", rw.Code)
	}
	if rw.Header().Get("Location") != "/hello/world" {
		t.Error("Test Handle failed", rw.Header())
	}

	r = httptest.NewRequest("GET", "/hello/world/elvizlai/", nil)
	rw = httptest.NewRecorder()
	router.ServeHTTP(rw, r)
	if rw.Code != http.StatusMovedPermanently {
		t.Error("Test Handle failed", rw.Code)
	}
	if rw.Header().Get("Location") != "/hello/world/elvizlai" {
		t.Error("Test Handle failed", rw.Header())
	}
}

func TestMiddleware(t *testing.T) {
	router := New()

	router.UseFunc(func() negroni.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			rw.Header().Set("m1", "m1")
			if rw.Header().Get("m2") != "" {
				t.Fatal("m2 should run after m1")
			}
			next(rw, r)
		}
	}())

	router.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		if rw.Header().Get("m1") != "m1" {
			t.Fatal("m1 should run first")
		}
		rw.Header().Set("m2", "m2")
		next(rw, r)
	})

	router.Handler("GET", "/hello/world", func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.WriteHeader(http.StatusTeapot)
	})

	r := httptest.NewRequest("GET", "/hello/world", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, r)
	if rw.Code != http.StatusTeapot {
		t.Error("TestMiddleware failed", rw.Code)
	}

	if rw.Header().Get("m1") == "" {
		t.Fatal("header m1 should exist")
	}

	if rw.Header().Get("m2") == "" {
		t.Fatal("header m2 should exist")
	}
}

func TestGroup(t *testing.T) {
	router := New()

	router.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Set("common", "x")
		next(rw, r)
	})

	v1 := router.Group("/v1", func() negroni.HandlerFunc {
		return func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
			rw.Header().Set("v1", "v1")
		}
	}())

	v2 := router.Group("/v2")
	v2.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Header().Set("v2", "v2")
		next(rw, r)
	})

	v1.GET("/test", func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Write([]byte("v1"))
	})

	v2.GET("/test", func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.Write([]byte("v2"))
	})

	r1 := httptest.NewRequest("GET", "/v1/test", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, r1)

	if rw.Header().Get("common") != "x" {
		t.Fatal("v1 header common should exist")
	}

	if rw.Header().Get("v1") == "" {
		t.Fatal("v1 header v1 should exist")
	}

	if rw.Header().Get("v2") != "" {
		t.Fatal("v1 header v2 should not exist")
	}

	r2 := httptest.NewRequest("GET", "/v2/test", nil)
	rw = httptest.NewRecorder()
	router.ServeHTTP(rw, r2)

	if rw.Header().Get("common") != "x" {
		t.Fatal("v2 header common should exist")
	}

	if rw.Header().Get("v1") != "" {
		t.Fatal("v2 header v1 should not exist")
	}

	if rw.Header().Get("v2") == "" {
		t.Fatal("v2 header v2 should exist")
	}
}

func TestNotFound(t *testing.T) {
	router := New()

	router.NotFound(func(rw http.ResponseWriter, r *http.Request) {
		rw.Header().Set("lol", "lol")
		rw.WriteHeader(http.StatusNotFound)
	})

	r := httptest.NewRequest("GET", "/v1/test", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, r)

	if rw.Code != http.StatusNotFound {
		t.Fatal("should return code 404")
	}

	if rw.Header().Get("lol") != "lol" {
		t.Fatal("header lol should exist")
	}

}

func TestPanic(t *testing.T) {
	router := New()

	router.PanicHandler(func(rw http.ResponseWriter, r *http.Request, i interface{}) {
		rw.Header().Set("panic-info", fmt.Sprint(i))
		rw.WriteHeader(http.StatusInternalServerError)
	})

	router.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		next(rw, r)
		rw.Header().Set("after", "after") // this should never happen, because m2 is panic
	})

	router.UseFunc(func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		panic("middleware 2 panic")
		next(rw, r)
	})

	router.GET("/panic", func(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		rw.WriteHeader(http.StatusOK)
		next(rw, r)
	})

	r := httptest.NewRequest("GET", "/panic", nil)
	rw := httptest.NewRecorder()

	router.ServeHTTP(rw, r)

	if rw.Code != http.StatusInternalServerError {
		t.Fatal("should return code 500")
	}

	if rw.Header().Get("panic-info") == "" {
		t.Fatal("panic-info should exist")
	}

	if rw.Header().Get("after") != "" {
		t.Fatal("after should exist")
	}
}
