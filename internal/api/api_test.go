package api_test

import (
	"testing"

	"git.sr.ht/~jakintosh/coffer/internal/api"
	"git.sr.ht/~jakintosh/coffer/internal/service"
	"git.sr.ht/~jakintosh/coffer/internal/util"
	"git.sr.ht/~jakintosh/coffer/pkg/wire"
)

func setupCORS(
	t *testing.T,
	env *util.TestEnv,
) {
	origins := []service.AllowedOrigin{
		{URL: "http://test-default"},
	}
	if err := env.Service.SetAllowedOrigins(origins); err != nil {
		t.Fatalf("failed to set cors: %v", err)
	}
}

func setupTestEnv(
	t *testing.T,
) *util.TestEnv {
	t.Helper()
	env := util.SetupTestEnv(t)
	api := api.New(env.Service)
	env.Router = api.BuildRouter()
	return env
}

func makeTestAuthHeader(
	t *testing.T,
	env *util.TestEnv,
) wire.TestHeader {
	token, err := env.Service.KeysService().Create()
	if err != nil {
		t.Fatal(err)
	}
	return wire.TestHeader{Key: "Authorization", Value: "Bearer " + token}
}
