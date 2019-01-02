package api

import (
	"context"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-gorp/gorp"
	"github.com/gorilla/mux"

	"github.com/ovh/cds/engine/api/auth"
	"github.com/ovh/cds/engine/api/bootstrap"
	"github.com/ovh/cds/engine/api/event"
	"github.com/ovh/cds/engine/api/test"
)

func newTestAPI(t *testing.T, bootstrapFunc ...test.Bootstrapf) (*API, *gorp.DbMap, *Router, context.CancelFunc) {
	bootstrapFunc = append(bootstrapFunc, bootstrap.InitiliazeDB)
	db, cache, end := test.SetupPG(t, bootstrapFunc...)
	router := newRouter(auth.TestLocalAuth(t, db), mux.NewRouter(), "/"+test.GetTestName(t))
	var cancel context.CancelFunc
	router.Background, cancel = context.WithCancel(context.Background())
	api := &API{
		StartupTime:         time.Now(),
		Router:              router,
		DBConnectionFactory: test.DBConnectionFactory,
		Config:              Configuration{},
		Cache:               cache,
	}
	_ = event.Initialize(event.KafkaConfig{}, api.Cache)
	api.Config.Auth.AuthenticationConfig.SigningKey = []byte("this is key")
	api.Config.Auth.Local.Enable = true
	api.InitRouter()
	f := func() {
		cancel()
		end()
	}
	return api, db, router, f
}

func newTestServer(t *testing.T, bootstrapFunc ...test.Bootstrapf) (*API, string, func()) {
	bootstrapFunc = append(bootstrapFunc, bootstrap.InitiliazeDB)
	db, cache, end := test.SetupPG(t, bootstrapFunc...)
	router := newRouter(auth.TestLocalAuth(t, db), mux.NewRouter(), "")
	var cancel context.CancelFunc
	router.Background, cancel = context.WithCancel(context.Background())
	api := &API{
		StartupTime:         time.Now(),
		Router:              router,
		DBConnectionFactory: test.DBConnectionFactory,
		Config:              Configuration{},
		Cache:               cache,
	}
	api.Config.Auth.AuthenticationConfig.SigningKey = []byte("this is key")
	api.Config.Auth.Local.Enable = true
	_ = event.Initialize(event.KafkaConfig{}, api.Cache)
	api.InitRouter()
	ts := httptest.NewServer(router.Mux)
	url, _ := url.Parse(ts.URL)
	f := func() {
		end()
		cancel()
		ts.Close()
	}
	return api, url.String(), f
}
