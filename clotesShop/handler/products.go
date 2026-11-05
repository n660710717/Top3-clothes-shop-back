package handler

import (
	"errors"
	"net/http"

	"clotesShop/store"
)

func (h *Handler) ListProducts(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListProducts(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) GetProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขสินค้าต้องเป็นตัวเลข")
		return
	}
	item, err := h.Store.ProductByID(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "ไม่พบสินค้านี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขสินค้าต้องเป็นตัวเลข")
		return
	}
	err := h.Store.DeleteProduct(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "ไม่พบสินค้านี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
