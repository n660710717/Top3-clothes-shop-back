package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"clotesShop/mw"
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

func (h *Handler) CreateSellerApplication(w http.ResponseWriter, r *http.Request) {
	var body struct {
		ShopName string `json:"shopName"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	name := strings.TrimSpace(body.ShopName)
	if name == "" || utf8.RuneCountInString(name) > 80 {
		writeError(w, http.StatusBadRequest, "BAD_SHOP", "ต้องมีชื่อร้าน และยาวไม่เกิน 80 ตัวอักษร")
		return
	}
	who := mw.WhoFrom(r.Context())
	err := h.Store.CreateSellerApplication(r.Context(), who.ID, name)
	if errors.Is(err, store.ErrAlreadyApplied) {
		writeError(w, http.StatusConflict, "ALREADY_APPLIED", "คุณมีคำขอที่รออนุมัติอยู่แล้ว")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true})
}

func (h *Handler) ListSellerApplications(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListSellerApplications(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ApproveSellerApplication(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขคำขอต้องเป็นตัวเลข")
		return
	}
	err := h.Store.ApproveSellerApplication(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "APPLICATION_NOT_FOUND", "ไม่พบคำขอนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
