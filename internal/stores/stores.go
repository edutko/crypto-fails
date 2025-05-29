package stores

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/edutko/crypto-fails/internal/job"
	"github.com/edutko/crypto-fails/internal/share"
	"github.com/edutko/crypto-fails/internal/store/blob"
	"github.com/edutko/crypto-fails/internal/store/encrypted"
	"github.com/edutko/crypto-fails/internal/store/kv"
	"github.com/edutko/crypto-fails/internal/user"
)

func Initialize(storageRootDir string) error {
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

	backupDir, err = os.OpenRoot(filepath.Join(storageRootDir, "tmp"))
	if err != nil {
		Cleanup()
		return err
	}

	fileStore, err = encryptedFileStore(storageRootDir)
	if err != nil {
		Cleanup()
		return err
	}

	jobStore = kv.NewMemoryStore[job.Descriptor]()

	shareStore, err = kv.Open[share.Link](filepath.Join(storageRootDir, "shares"))
	if err != nil {
		Cleanup()
		return err
	}

	userStore, err = kv.Open[user.User](filepath.Join(storageRootDir, "users"))
	if err != nil {
		Cleanup()
		return err
	}

	return nil
}

func BackupDir() *os.Root {
	return backupDir
}

func FileStore() *encrypted.ObjectStore {
	return fileStore
}

func JobStore() kv.Store[job.Descriptor] {
	return jobStore
}

func ShareStore() kv.Store[share.Link] {
	return shareStore
}

func UserStore() kv.Store[user.User] {
	return userStore
}

func Cleanup() {
	if backupDir != nil {
		_ = backupDir.Close()
	}

	if fileStore != nil {
		_ = fileStore.Close()
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

	fileStore = nil
	jobStore = nil
	shareStore = nil
	userStore = nil
}

func encryptedFileStore(rootDir string) (*encrypted.ObjectStore, error) {
	blobStore, err := blob.Open(filepath.Join(rootDir, "files"))
	if err != nil {
		return nil, err
	}

	keyStore, err := kv.Open[[]byte](filepath.Join(rootDir, "keys"))
	if err != nil {
		_ = blobStore.Close()
		return nil, err
	}

	fs, err := encrypted.NewObjectStore(blobStore, keyStore)
	if err != nil {
		_ = blobStore.Close()
		_ = keyStore.Close()
	}
	return fs, err
}

var backupDir *os.Root
var fileStore *encrypted.ObjectStore
var jobStore *kv.MemoryStore[job.Descriptor]
var shareStore *kv.FSStore[share.Link]
var userStore *kv.FSStore[user.User]
