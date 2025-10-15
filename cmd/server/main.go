package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/duykhoa/gopass/internal/config"
	"github.com/duykhoa/gopass/internal/gpg"
	"github.com/duykhoa/gopass/internal/service"
	"github.com/duykhoa/gopass/internal/store"
)

func main() {
	http.HandleFunc("/", helloHandler)
	http.HandleFunc("/secrets", listSecretsHandler)
	http.HandleFunc("/secrets/", secretHandler)
	http.HandleFunc("/init", initHandler)

	fmt.Println("Server is listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Welcome to gopass HTTP API")
}

func listSecretsHandler(w http.ResponseWriter, r *http.Request) {
	storeDir := config.PasswordStoreDir()
	files, err := os.ReadDir(storeDir)
	if err != nil {
		http.Error(w, "Failed to read password store", http.StatusInternalServerError)
		return
	}

	var secrets []string
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".gpg") {
			secrets = append(secrets, strings.TrimSuffix(file.Name(), ".gpg"))
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(secrets)
}

func secretHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		viewSecret(w, r)
	case http.MethodPut:
		updateSecret(w, r)
	case http.MethodDelete:
		deleteSecret(w, r)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func viewSecret(w http.ResponseWriter, r *http.Request) {
	secretName := strings.TrimPrefix(r.URL.Path, "/secrets/")
	passphrase := r.Header.Get("X-Gopass-Passphrase")

	if passphrase == "" {
		http.Error(w, "X-Gopass-Passphrase header is required", http.StatusBadRequest)
		return
	}

	req := service.DecryptRequest{
		StoreDir:   config.PasswordStoreDir(),
		Entry:      secretName,
		Passphrase: passphrase,
		CachePath:  service.GetDefaultCachePath(),
	}

	result := service.Decrypt(req)
	if result.Err != nil {
		http.Error(w, fmt.Sprintf("Failed to decrypt secret: %v", result.Err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"value": result.Plaintext})
}

func updateSecret(w http.ResponseWriter, r *http.Request) {
	secretName := strings.TrimPrefix(r.URL.Path, "/secrets/")
	gpgID := config.GPGId()

	if gpgID == "" {
		http.Error(w, "GPG ID is not configured", http.StatusInternalServerError)
		return
	}

	var requestBody map[string]string
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newValue, ok := requestBody["value"]
	if !ok {
		http.Error(w, "Missing 'value' in request body", http.StatusBadRequest)
		return
	}

	encryptedValue, err := gpg.EncryptWithGPGKey([]byte(newValue), gpgID)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to encrypt secret: %v", err), http.StatusInternalServerError)
		return
	}

	storeDir := config.PasswordStoreDir()
	secretPath := filepath.Join(storeDir, secretName+".gpg")

	err = os.WriteFile(secretPath, encryptedValue, 0600)
	if err != nil {
		http.Error(w, "Failed to write secret to store", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func deleteSecret(w http.ResponseWriter, r *http.Request) {
	secretName := strings.TrimPrefix(r.URL.Path, "/secrets/")
	storeDir := config.PasswordStoreDir()
	secretPath := filepath.Join(storeDir, secretName+".gpg")

	err := os.Remove(secretPath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "Secret not found", http.StatusNotFound)
		} else {
			http.Error(w, "Failed to delete secret", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func initHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var requestBody map[string]string
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	gitRepoURL, ok := requestBody["git_repo_url"]
	if !ok {
		http.Error(w, "Missing 'git_repo_url' in request body", http.StatusBadRequest)
		return
	}

	gpgKey, ok := requestBody["gpg_key"]
	if !ok {
		http.Error(w, "Missing 'gpg_key' in request body", http.StatusBadRequest)
		return
	}

	storeDir := config.PasswordStoreDir()

	err = store.InitPasswordStore(storeDir, gpgKey, gitRepoURL)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to initialize password store: %v", err), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
