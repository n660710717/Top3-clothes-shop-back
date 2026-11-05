package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"clotesShop/store"
)

func (h *Handler) AdminMembers(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListMembers(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

type roleReq struct {
	Role string `json:"role"`
}

func (h *Handler) SetMemberRole(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "memberId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขสมาชิกต้องเป็นตัวเลข")
		return
	}
	var req roleReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	switch req.Role {
	case "buyer", "seller", "admin":
	default:
		writeError(w, http.StatusBadRequest, "BAD_ROLE", "บทบาทต้องเป็น buyer, seller หรือ admin")
		return
	}
	m, err := h.Store.SetMemberRole(r.Context(), id, req.Role)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "MEMBER_NOT_FOUND", "ไม่พบสมาชิกนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, m)
}
