package route

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/edutko/crypto-fails/internal/job"
	"github.com/edutko/crypto-fails/internal/middleware"
	"github.com/edutko/crypto-fails/internal/route/responses"
	"github.com/edutko/crypto-fails/internal/store"
	"github.com/edutko/crypto-fails/internal/stores"
	"github.com/edutko/crypto-fails/pkg/blob"
)

func PostBackups(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)

	if id, err := startBackup(s.Username); err != nil {
		responses.InternalServerError(w, err)
	} else {
		responses.Created(w, path.Join("api", "backups", url.PathEscape(id), "status"))
	}
}

func GetBackup(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	id := path.Clean(r.PathValue("id"))
	if !strings.HasPrefix(id, s.Username+"/") {
		responses.Forbidden(w, fmt.Errorf("access denied for %q on %q", s.Username, r.PathValue("id")))
		return
	}

	j, err := stores.JobStore().Get(id)
	if errors.Is(err, store.ErrNotFound) {
		responses.NotFound(w)
		return
	} else if err != nil {
		responses.InternalServerError(w, err)
		return
	}

	if j.Status() == job.StatusCompleted {
		name := jobIdToFilename(j.Id)
		if fr, err := stores.BackupDir().Open(name); errors.Is(err, fs.ErrNotExist) {
			responses.NotFound(w)
		} else if err != nil {
			responses.InternalServerError(w, err)
		} else {
			responses.DownloadFromReader(w, name, fr)
		}
	} else {
		responses.SeeOther(w, path.Join("api", "jobs", id))
	}
}

func DeleteBackup(w http.ResponseWriter, r *http.Request) {
	s := middleware.GetCurrentSession(r)
	id := path.Clean(r.PathValue("id"))
	if !strings.HasPrefix(id, s.Username+"/") {
		responses.Forbidden(w, fmt.Errorf("access denied for %q on %q", s.Username, r.PathValue("id")))
		return
	}

	_, err := stores.JobStore().Delete(id)
	if err == nil || errors.Is(err, store.ErrNotFound) {
		responses.NoContent(w)
	} else {
		responses.InternalServerError(w, err)
	}
}

func startBackup(uid string) (string, error) {
	j := job.Descriptor{
		Id:        path.Join(uid, time.Now().Format("20060102-1504")),
		StartedAt: time.Now().UTC(),
	}
	if err := stores.JobStore().Put(j.Id, j); err != nil {
		return "", err
	}

	files, err := stores.FileStore().ListObjectsWithPrefix(uid)
	if err != nil {
		return "", err
	}

	go createZip(uid, j, files)

	return j.Id, nil
}

func createZip(username string, j job.Descriptor, files []blob.Metadata) {
	zf, err := stores.BackupDir().Create(jobIdToFilename(j.Id))
	if err != nil {
		log.Println(err)
		return
	}
	defer zf.Close()

	zw := zip.NewWriter(zf)
	defer zw.Close()

	for _, m := range files {
		err = zipFile(zw, m.Key, strings.TrimPrefix(m.Key, username+"/"))
		if err != nil {
			log.Println(err)
			continue
		}
	}

	j.FinishedAt = time.Now().UTC()
	if err = stores.JobStore().Put(j.Id, j); err != nil {
		log.Println(err)
	}
}

func zipFile(zw *zip.Writer, key, zipPath string) error {
	r, _, err := stores.FileStore().GetObject(key)
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := zw.Create(zipPath)
	if err != nil {
		return err
	}

	_, err = io.Copy(w, r)

	return err
}

func jobIdToFilename(id string) string {
	return strings.ReplaceAll(id, "/", "-") + ".zip"
}
