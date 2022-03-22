package coderd

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/coder/coder/codersdk"
	"github.com/coder/coder/database"
	"github.com/coder/coder/httpapi"
	"github.com/coder/coder/httpmw"
)

func (api *api) postFile(rw http.ResponseWriter, r *http.Request) {
	apiKey := httpmw.APIKey(r)
	contentType := r.Header.Get("Content-Type")

	switch contentType {
	case "application/x-tar":
	default:
		httpapi.Write(rw, http.StatusBadRequest, httpapi.Response{
			Message: fmt.Sprintf("unsupported content type: %s", contentType),
		})
		return
	}

	r.Body = http.MaxBytesReader(rw, r.Body, 10*(10<<20))
	data, err := io.ReadAll(r.Body)
	if err != nil {
		httpapi.Write(rw, http.StatusBadRequest, httpapi.Response{
			Message: fmt.Sprintf("read file: %s", err),
		})
		return
	}
	hashBytes := sha256.Sum256(data)
	hash := hex.EncodeToString(hashBytes[:])
	file, err := api.Database.GetFileByHash(r.Context(), hash)
	if err == nil {
		// The file already exists!
		render.Status(r, http.StatusOK)
		render.JSON(rw, r, codersdk.UploadResponse{
			Hash: file.Hash,
		})
		return
	}
	file, err = api.Database.InsertFile(r.Context(), database.InsertFileParams{
		Hash:      hash,
		CreatedBy: apiKey.UserID,
		CreatedAt: database.Now(),
		Mimetype:  contentType,
		Data:      data,
	})
	if err != nil {
		httpapi.Write(rw, http.StatusInternalServerError, httpapi.Response{
			Message: fmt.Sprintf("insert file: %s", err),
		})
		return
	}
	render.Status(r, http.StatusCreated)
	render.JSON(rw, r, codersdk.UploadResponse{
		Hash: file.Hash,
	})
}

func (api *api) fileByHash(rw http.ResponseWriter, r *http.Request) {
	hash := chi.URLParam(r, "hash")
	if hash == "" {
		httpapi.Write(rw, http.StatusBadRequest, httpapi.Response{
			Message: "hash must be provided",
		})
		return
	}
	file, err := api.Database.GetFileByHash(r.Context(), hash)
	if errors.Is(err, sql.ErrNoRows) {
		httpapi.Write(rw, http.StatusNotFound, httpapi.Response{
			Message: "no file exists with that hash",
		})
		return
	}
	if err != nil {
		httpapi.Write(rw, http.StatusInternalServerError, httpapi.Response{
			Message: fmt.Sprintf("get file: %s", err),
		})
		return
	}
	rw.Header().Set("Content-Type", file.Mimetype)
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(file.Data)
}
