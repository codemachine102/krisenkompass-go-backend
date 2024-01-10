package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"bitbucket.org/ibros_nsk/krisenkompass-backend/internal/authorize"
	"bitbucket.org/ibros_nsk/krisenkompass-backend/pkg/randstring"
)

func (h *Handler) UploadFile(w http.ResponseWriter, r *http.Request) {
	path := r.FormValue("path")
	path = strings.Trim(path, "/")
	pathParts := strings.Split(path, "/")
	if len(path) != 0 {
		path = "../../storage/" + path + "/"
	}

	// Check if user has rights to access requested path.
	if pathParts[0] == "organization" {
		id, err := strconv.ParseInt(pathParts[1], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		_, err = authorize.AuthorizeOrganization(r.Context(), id, "user")
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	} else if pathParts[0] == "user" {
		id, err := strconv.ParseInt(pathParts[1], 10, 64)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		claims, err := authorize.Authorize(r.Context(), "user")
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		if claims.UserID != id {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	} else if pathParts[0] == "common" {
		_, err := authorize.Authorize(r.Context(), "superadmin")
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			return
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		err := os.MkdirAll(path, os.ModePerm)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// Parse our multipart form, 10 << 20 specifies a maximum upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		fmt.Println(err)
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer file.Close()

	extension := filepath.Ext(handler.Filename)
	filename := strings.TrimSuffix(handler.Filename, extension)
	// Add random string to the end if file with given name exist.
	if _, err = os.Stat(path + filename + extension); err == nil {
		filename = fmt.Sprintf("%s-%s", filename, randstring.RandAlphanumString(5))
	}

	filePath := fmt.Sprintf("%s%s%s", path, filename, extension)

	// Create a new file in the uploads directory
	dst, err := os.Create(filePath)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file to the filesystem at the specified destination.
	_, err = io.Copy(dst, file)
	if err != nil {
		fmt.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"path": strings.TrimPrefix(filePath, "../../storage"),
	})
}
