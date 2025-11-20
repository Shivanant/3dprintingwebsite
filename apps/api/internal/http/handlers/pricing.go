package handlers

import (
	"net/http"

	"github.com/3dprint-hub/api/internal/jobs"
	httpmw "github.com/3dprint-hub/api/internal/http/middleware"
)

func (h *Handler) EstimatePrice(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(25 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid form data")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file required")
		return
	}
	defer file.Close()

	estimate, data, err := h.App.Pricing.EstimateFromUpload(r.Context(), header)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	material := r.FormValue("material")
	if material == "" {
		material = "PLA"
	}
	quality := r.FormValue("quality")
	if quality == "" {
		quality = "standard"
	}

	if userCtx, ok := httpmw.GetUser(r.Context()); ok {
		_, err := h.App.Jobs.Create(r.Context(), jobs.CreateInput{
			UserID:   userCtx.UserID,
			FileName: header.Filename,
			Data:     data,
			Estimate: estimate,
			Material: material,
			Quality:  quality,
		})
		if err != nil {
			h.App.Logger.Warn("failed to persist print job", "err", err)
		}
	}

	writeJSON(w, http.StatusOK, estimate)
}
