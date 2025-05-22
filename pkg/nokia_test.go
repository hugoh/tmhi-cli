package pkg

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewNokiaGateway(t *testing.T) {
	url := "http://localhost"
	gw := NewNokiaGateway(url)

	if gw.baseURL != url {
		t.Errorf("expected baseURL %s, got %s", url, gw.baseURL)
	}

	if gw.httpClient == nil {
		t.Errorf("expected httpClient to be initialized, but it was nil")
	}
	// TODO: Check other default values if any, e.g. if there's a default timeout
}

func TestLogin(t *testing.T) {
	// Test case 1: Successful login
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login_app.cgi" && r.Method == http.MethodPost {
			http.SetCookie(w, &http.Cookie{Name: "session_id", Value: "test-session-id"})
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	gw := NewNokiaGateway(server.URL)
	err := gw.Login("admin", "password")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if gw.sessionID != "test-session-id" {
		t.Errorf("expected sessionID to be 'test-session-id', got '%s'", gw.sessionID)
	}

	// Test case 2: Failed login (incorrect credentials)
	serverFail := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login_app.cgi" && r.Method == http.MethodPost {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer serverFail.Close()

	gwFail := NewNokiaGateway(serverFail.URL)
	errFail := gwFail.Login("admin", "wrongpassword")
	if errFail == nil {
		t.Errorf("expected an error for failed login, got nil")
	}
	if gwFail.sessionID != "" {
		t.Errorf("expected sessionID to be empty for failed login, got '%s'", gwFail.sessionID)
	}

	// Test case 3: Network error
	gwNetErr := NewNokiaGateway("http://localhost:12345")
	errNetErr := gwNetErr.Login("admin", "password")
	if errNetErr == nil {
		t.Errorf("expected a network error, got nil")
	}
}

func TestReboot(t *testing.T) {
	// Test case 1: Successful reboot
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/reboot_app.cgi" && r.Method == http.MethodPost {
			cookie, err := r.Cookie("session_id")
			if err != nil || cookie.Value != "test-session-id" {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	gw := NewNokiaGateway(server.URL)
	gw.sessionID = "test-session-id" // Simulate logged-in state
	err := gw.Reboot(false)
	if err != nil {
		t.Errorf("expected no error for successful reboot, got %v", err)
	}

	// Test case 2: Failed reboot (not logged in)
	gwNotLoggedIn := NewNokiaGateway(server.URL)
	errNotLoggedIn := gwNotLoggedIn.Reboot(false)
	if errNotLoggedIn == nil {
		t.Errorf("expected an error when trying to reboot without being logged in, got nil")
	}

	// Test case 3: Dry run
	gwDryRun := NewNokiaGateway(server.URL)
	gwDryRun.sessionID = "test-session-id"
	errDryRun := gwDryRun.Reboot(true)
	if errDryRun != nil {
		t.Errorf("expected no error for dry run, got %v", errDryRun)
	}

	// Test case 4: Network error
	gwNetErrReboot := NewNokiaGateway("http://localhost:12345")
	gwNetErrReboot.sessionID = "test-session-id"
	errNetErrReboot := gwNetErrReboot.Reboot(false)
	if errNetErrReboot == nil {
		t.Errorf("expected a network error when rebooting with a non-existent server, got nil")
	}
}

func TestGetNonce(t *testing.T) {
	// Test case 1: Successful nonce retrieval
	expectedNonce := "12345abcdef"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login_app.cgi" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"nonce": "` + expectedNonce + `"}`))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	gw := NewNokiaGateway(server.URL)
	nonce, err := gw.getNonce()
	if err != nil {
		t.Errorf("expected no error for successful nonce retrieval, got %v", err)
	}
	if nonce != expectedNonce {
		t.Errorf("expected nonce %s, got %s", expectedNonce, nonce)
	}

	// Test case 2: Server returns an error
	serverError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login_app.cgi" && r.Method == http.MethodGet {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer serverError.Close()

	gwServerError := NewNokiaGateway(serverError.URL)
	_, errServerError := gwServerError.getNonce()
	if errServerError == nil {
		t.Errorf("expected an error when server returns 500, got nil")
	}

	// Test case 3: Non-JSON response or malformed JSON
	serverMalformed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/login_app.cgi" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`this is not json`))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer serverMalformed.Close()

	gwMalformed := NewNokiaGateway(serverMalformed.URL)
	_, errMalformed := gwMalformed.getNonce()
	if errMalformed == nil {
		t.Errorf("expected an error for malformed JSON response, got nil")
	}
	
	// Test case 4: Network error
	gwNetErrNonce := NewNokiaGateway("http://localhost:12345")
	_, errNetErrNonce := gwNetErrNonce.getNonce()
	if errNetErrNonce == nil {
		t.Errorf("expected a network error when getting nonce from a non-existent server, got nil")
	}
}

func TestGetCredentials(t *testing.T) {
	// Test case 1: Successful credential retrieval
	// This test assumes getCredentials retrieves from a source that can be mocked
	// or is statically configured within the NokiaGateway for testing.
	// If getCredentials involves an HTTP call to a predefined internal endpoint:
	expectedUser := "testuser"
	expectedPass := "testpass"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assuming getCredentials fetches from a path like "/api/internal-credentials"
		if r.URL.Path == "/api/internal-credentials" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"username": "` + expectedUser + `", "password": "` + expectedPass + `"}`))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer server.Close()

	// Create a gateway. If getCredentials uses baseURL, it should be set to server.URL.
	// If getCredentials reads directly from env or a global config, the mock server might not be used.
	gw := NewNokiaGateway(server.URL) // Or NewNokiaGateway("") if baseURL is not used by getCredentials

	user, pass, err := gw.getCredentials() // Assumes no arguments
	if err != nil {
		t.Errorf("expected no error for successful credential retrieval, got %v", err)
	}
	if user != expectedUser {
		t.Errorf("expected username %s, got %s", expectedUser, user)
	}
	if pass != expectedPass {
		t.Errorf("expected password %s, got %s", expectedPass, pass)
	}

	// Test case 2: Server returns an error (if getCredentials makes an HTTP call)
	serverError := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/internal-credentials" && r.Method == http.MethodGet {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer serverError.Close()

	gwServerError := NewNokiaGateway(serverError.URL)
	_, _, errServerError := gwServerError.getCredentials()
	if errServerError == nil {
		t.Errorf("expected an error when server returns 500 for credentials, got nil")
	}
	// This assertion depends on getCredentials actually making a call to serverError.URL
	// and correctly propagating the HTTP error.

	// Test case 3: Malformed JSON response (if getCredentials makes an HTTP call and parses JSON)
	serverMalformed := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/internal-credentials" && r.Method == http.MethodGet {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`this is not json`))
			return
		}
		http.Error(w, "not found", http.StatusNotFound)
	}))
	defer serverMalformed.Close()

	gwMalformed := NewNokiaGateway(serverMalformed.URL)
	_, _, errMalformed := gwMalformed.getCredentials()
	if errMalformed == nil {
		t.Errorf("expected an error for malformed JSON credentials response, got nil")
	}

	// Test case 4: Network error (if getCredentials makes an HTTP call)
	// This test is relevant if getCredentials makes an HTTP request.
	gwNetErrCreds := NewNokiaGateway("http://localhost:12345") // Non-existent server
	_, _, errNetErrCreds := gwNetErrCreds.getCredentials()
	if errNetErrCreds == nil {
		t.Errorf("expected a network error when getting credentials from a non-existent server, got nil")
	}
	// Add a placeholder for non-HTTP getCredentials:
	// if getCredentials is supposed to get from env vars, test that too.
	// e.g. t.Run("credentials from env", func(t *testing.T) { os.Setenv(...); ...})
}
