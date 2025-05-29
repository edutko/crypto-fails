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

	"github.com/edutko/crypto-fails/internal/auth"
	"github.com/edutko/crypto-fails/internal/config"
	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/info"
	m "github.com/edutko/crypto-fails/internal/middleware"
	"github.com/edutko/crypto-fails/internal/route"
	"github.com/edutko/crypto-fails/internal/store"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/internal/user"
	"github.com/edutko/crypto-fails/internal/user/role"
)

func main() {
	info.Initialize(Version)
	conf := config.Load()

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

	mux.HandleFunc("/login", route.LoginUI)
	mux.HandleFunc("/logout", route.Logout)
	mux.HandleFunc("/register", route.Register)

	mux.HandleFunc("GET /{$}", m.MaybeAuthenticated(route.Index))
	mux.HandleFunc("GET /admin", m.RequireAdmin(route.Admin))
	mux.HandleFunc("GET /download", m.MaybeAuthenticated(route.Download))
	mux.HandleFunc("GET /files", m.Authenticated(route.MyFiles))
	mux.HandleFunc("GET /keys/{id}", route.Pubkey)
	mux.HandleFunc("GET /shares", m.Authenticated(route.MyShares))
	mux.HandleFunc("POST /share", m.Authenticated(route.NewShare))
	mux.HandleFunc("POST /upload", m.Authenticated(route.Upload))

	if conf.LeakEncryptedFiles {
		fs := http.Dir(filepath.Join(conf.StorageRootDir, "files"))
		mux.Handle("GET /vulns/leak/", http.StripPrefix("/vulns/leak/", http.FileServer(fs)))
	}
	if conf.TweakEncryptedFiles {
		mux.Handle("PUT /vulns/tweak/{key...}", http.StripPrefix("/vulns/tweak/", http.HandlerFunc(route.TweakCiphertext)))
	}

	mux.HandleFunc("/api/login", route.LoginAPI)
	mux.HandleFunc("/api/logout", route.Logout)

	mux.HandleFunc("/api/backups", m.Authenticated(route.Backups))
	mux.HandleFunc("/api/backups/{id}", m.Authenticated(route.Backup))
	mux.HandleFunc("/api/files", m.Authenticated(route.Files))
	mux.HandleFunc("/api/files/{key}", m.Authenticated(route.File))
	mux.HandleFunc("/api/jobs/{id}", m.Authenticated(route.Job))
	mux.HandleFunc("/api/keys", m.Authenticated(route.Pubkeys))
	mux.HandleFunc("/api/keys/{id}", m.MaybeAuthenticated(route.Pubkey))
	mux.HandleFunc("/api/shares", m.Authenticated(route.Shares))
	mux.HandleFunc("/api/shares/{id}", m.Authenticated(route.Share))
	mux.HandleFunc("/api/users", m.RequireAdmin(route.Users))
	mux.HandleFunc("/api/users/{id}/keys", m.Authenticated(route.UserPubkeys))

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
