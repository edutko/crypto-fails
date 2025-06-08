package stores

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/edutko/crypto-fails/internal/crypto"
	"github.com/edutko/crypto-fails/internal/job"
	"github.com/edutko/crypto-fails/internal/store/blob"
	"github.com/edutko/crypto-fails/internal/store/encrypted"
	"github.com/edutko/crypto-fails/internal/store/kv"
	"github.com/edutko/crypto-fails/pkg/share"
	"github.com/edutko/crypto-fails/pkg/user"
)

func Initialize(storageRootDir string, encryptionMode crypto.Mode) error {
	if err := os.MkdirAll(storageRootDir, 0755); err != nil {
		return err
	}

	storageRoot, err := os.OpenRoot(storageRootDir)
	if err != nil {
		return err
	}
	defer storageRoot.Close()

	if err = storageRoot.Mkdir("files", 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	if err = os.RemoveAll(filepath.Join(storageRootDir, "tmp")); err != nil {
		return err
	}
	if err = storageRoot.Mkdir("tmp", 0755); err != nil && !errors.Is(err, os.ErrExist) {
		return err
	}

	if backupRoot, err = os.OpenRoot(filepath.Join(storageRootDir, "tmp")); err != nil {
		Cleanup()
		return err
	}

	if blobStore, err = blob.Open(filepath.Join(storageRootDir, "files")); err != nil {
		Cleanup()
		return err
	}

	if keyStore, err = kv.Open[[]byte](filepath.Join(storageRootDir, "keys")); err != nil {
		Cleanup()
		return err
	}

	if fileStore, err = encrypted.NewObjectStore(blobStore, keyStore, encryptionMode); err != nil {
		Cleanup()
		return err
	}

	jobStore = kv.NewInMemoryStore[job.Descriptor]()

	if shareStore, err = kv.Open[share.Link](filepath.Join(storageRootDir, "shares")); err != nil {
		Cleanup()
		return err
	}

	if userStore, err = kv.Open[user.User](filepath.Join(storageRootDir, "users")); err != nil {
		Cleanup()
		return err
	}

	return nil
}

func BackupDir() *os.Root {
	return backupRoot
}

func FileStore() *encrypted.ObjectStore {
	return fileStore
}

func JobStore() kv.Store[job.Descriptor] {
	return jobStore
}

func KeyStore() kv.Store[[]byte] {
	return keyStore
}

func ShareStore() kv.Store[share.Link] {
	return shareStore
}

func UserStore() kv.Store[user.User] {
	return userStore
}

func Cleanup() {
	if backupRoot != nil {
		_ = backupRoot.Close()
	}

	if fileStore != nil {
		_ = fileStore.Close()

	} else {
		if blobStore != nil {
			_ = blobStore.Close()
		}
		if keyStore != nil {
			_ = keyStore.Close()
		}
	}

	if jobStore != nil {
		_ = jobStore.Close()
	}

	if shareStore != nil {
		_ = shareStore.Close()
	}

	if userStore != nil {
		_ = userStore.Close()
	}

	backupRoot = nil
	blobStore = nil
	fileStore = nil
	jobStore = nil
	keyStore = nil
	shareStore = nil
	userStore = nil
}

var backupRoot *os.Root
var blobStore blob.Store
var fileStore *encrypted.ObjectStore
var jobStore kv.Store[job.Descriptor]
var keyStore kv.Store[[]byte]
var shareStore kv.Store[share.Link]
var userStore kv.Store[user.User]
