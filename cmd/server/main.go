package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/edutko/crypto-fails/internal/app"
	"github.com/edutko/crypto-fails/internal/auth"
	"github.com/edutko/crypto-fails/internal/config"
	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/crypto/random"
	m "github.com/edutko/crypto-fails/internal/middleware"
	"github.com/edutko/crypto-fails/internal/route"
	"github.com/edutko/crypto-fails/internal/store"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/pkg/user"
	"github.com/edutko/crypto-fails/pkg/user/role"
)

func main() {
	app.SetVersion(Version)
	conf := config.Load()

	random.SetWeakPRNG(conf.WeakPRNGAlgorithm)

	if err := stores.Initialize(conf.StorageRootDir, conf.FileEncryptionMode); err != nil {
		log.Fatal(err)
	}
	defer stores.Cleanup()

	if err := auth.InitializeKeys(); err != nil {
		log.Fatal(err)
	}

	ensureAdminUserExists()

	staticFilesRoot, err := os.OpenRoot(conf.WebRootDir)
	if err != nil {
		log.Fatal(err)
	}
	defer staticFilesRoot.Close()
	fileServer := http.FileServerFS(staticFilesRoot.FS())

	mux := http.NewServeMux()
	mux.Handle("GET /favicon.ico", fileServer)
	mux.Handle("GET /static/", http.StripPrefix("/static/", fileServer))
	mux.Handle("GET /.well-known/jwks.json", http.HandlerFunc(route.JWKS))

	mux.HandleFunc("GET /register", route.GetRegister)
	mux.HandleFunc("POST /register", route.PostRegister)

	mux.HandleFunc("GET /login", route.GetLoginUI)
	mux.HandleFunc("POST /login", route.PostLoginUI)
	mux.HandleFunc("/logout", route.Logout)

	mux.HandleFunc("GET /forgot-password", route.GetForgotPassword)
	mux.HandleFunc("POST /forgot-password", route.PostForgotPassword)
	mux.HandleFunc("GET /reset-password", route.GetForgotPassword)
	mux.HandleFunc("POST /reset-password", route.PostForgotPassword)

	mux.HandleFunc("GET /{$}", m.MaybeAuthenticated(route.GetIndex))
	mux.HandleFunc("GET /admin", m.RequireAdmin(route.GetAdmin))
	mux.HandleFunc("GET /download", m.MaybeAuthenticated(route.GetDownload))
	mux.HandleFunc("GET /files", m.Authenticated(route.GetMyFiles))
	mux.HandleFunc("GET /keys/{id}", route.GetPubkey)
	mux.HandleFunc("GET /shares", m.Authenticated(route.GetMyShares))
	mux.HandleFunc("POST /share", m.Authenticated(route.PostShare))
	mux.HandleFunc("POST /upload", m.Authenticated(route.PostUpload))

	if conf.LeakEncryptedFiles {
		fs := http.Dir(filepath.Join(conf.StorageRootDir, "files"))
		mux.Handle("GET /vulns/leak/", http.StripPrefix("/vulns/leak/", http.FileServer(fs)))
	}
	if conf.TweakEncryptedFiles {
		mux.Handle("PUT /vulns/tweak/{key...}", http.StripPrefix("/vulns/tweak/", http.HandlerFunc(route.PutCiphertext)))
	}

	mux.HandleFunc("POST /api/login", route.PostLoginAPI)
	mux.HandleFunc("/api/logout", route.Logout)

	mux.HandleFunc("POST /api/backups", m.Authenticated(route.PostBackups))
	mux.HandleFunc("GET /api/backups/{id}", m.Authenticated(route.GetBackup))
	mux.HandleFunc("DELETE /api/backups/{id}", m.Authenticated(route.DeleteBackup))

	mux.HandleFunc("GET /api/files", m.Authenticated(route.GetFiles))
	mux.HandleFunc("POST /api/files", m.Authenticated(route.PostFiles))
	mux.HandleFunc("GET /api/files/{key}", m.Authenticated(route.GetFile))
	mux.HandleFunc("DELETE /api/files/{key}", m.Authenticated(route.DeleteFile))

	mux.HandleFunc("GET /api/jobs/{id}", m.Authenticated(route.GetJob))

	mux.HandleFunc("GET /api/keys", m.Authenticated(route.GetPubkeys))
	mux.HandleFunc("POST /api/keys", m.Authenticated(route.PostPubkeys))
	mux.HandleFunc("GET /api/keys/{id}", m.MaybeAuthenticated(route.GetPubkey))

	mux.HandleFunc("GET /api/shares", m.Authenticated(route.GetShares))
	mux.HandleFunc("POST /api/shares", m.Authenticated(route.PostShares))
	mux.HandleFunc("DELETE /api/shares/{id}", m.Authenticated(route.DeleteShare))

	mux.HandleFunc("GET /api/users", m.RequireAdmin(route.GetUsers))
	mux.HandleFunc("POST /api/users", m.RequireAdmin(route.PostUsers))
	mux.HandleFunc("GET /api/users/{username}/keys", m.Authenticated(route.GetUserPubkeys))

	serve(conf.ListenAddr, mux)
}

func serve(addr string, mux *http.ServeMux) {
	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}
	go func() {
		fmt.Printf("Listening on %s...\n", addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	shutdownCtx, shutdownRelease := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownRelease()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func ensureAdminUserExists() {
	_, err := stores.UserStore().Get(defaultAdminUsername)
	if errors.Is(err, store.ErrNotFound) {
		ph, err := crypto.HashPassword(initialAdminPassword)
		if err != nil {
			panic(err)
		}

		if err = stores.UserStore().Put(defaultAdminUsername, user.User{
			Username:     defaultAdminUsername,
			PasswordHash: ph,
			Roles:        []string{role.Admin},
			RealName:     "Default Administrator",
		}); err != nil {
			panic(err)
		}

	} else if err != nil {
		panic(err)
	}
}

const (
	defaultAdminUsername = "admin"
	initialAdminPassword = "admin"
)

var Version = "0.0.0"
