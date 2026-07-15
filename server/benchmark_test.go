package internal

import (
	"testing"
)

func BenchmarkPageHomepage(b *testing.B) {
	app := testSetupApp(b)
	ts := startTestServer(b, app)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := ts.Client.Get("/")
		if err == nil {
			resp.Close()
		}
	}
}

func BenchmarkPageDocs(b *testing.B) {
	app := testSetupApp(b)
	ts := startTestServer(b, app)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := ts.Client.Get("/docs")
		if err == nil {
			resp.Close()
		}
	}
}

func BenchmarkPageTerms(b *testing.B) {
	app := testSetupApp(b)
	ts := startTestServer(b, app)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := ts.Client.Get("/terms")
		if err == nil {
			resp.Close()
		}
	}
}

func BenchmarkAPILogin(b *testing.B) {
	app := testSetupApp(b)
	ts := startTestServer(b, app)

	// Pre-create the user
	respReg, err := ts.Client.R().
		SetFormDataWithMap(map[string]string{
			"moniker":          "Bench User",
			"name":             "bench_user",
			"email":            "bench@example.com",
			"phone":            "9999999999",
			"password":         "password123",
			"confirm-password": "password123",
		}).
		Post("/account")
	if err == nil {
		respReg.Close()
	}

	reqLogin := ts.Client.R().
		SetFormDataWithMap(map[string]string{
			"email":    "bench@example.com",
			"password": "password123",
		})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := reqLogin.Post("/account/login")
		if err == nil {
			resp.Close()
		}
	}
}

func BenchmarkAPIPostMessage(b *testing.B) {
	app := testSetupApp(b)
	ts := startTestServer(b, app)

	// Pre-create user & group
	cookie, user := createTestUserCookie(b, ts, "Bench Sender", "bench_sender", "bench_sender@example.com")
	_ = createTestGroupDirectly(b, int64(user.Id), "Bench Group", "bench_group")

	reqMsg := ts.Client.R().
		SetHeader("Cookie", cookie).
		SetFormDataWithMap(map[string]string{
			"content": "Benchmarking message post",
		})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		resp, err := reqMsg.Post("/groups/1/messages/create")
		if err == nil {
			resp.Close()
		}
	}
}
