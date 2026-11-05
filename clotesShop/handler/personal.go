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

func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขสินค้าต้องเป็นตัวเลข")
		return
	}
	items, err := h.Store.ListReviews(r.Context(), id)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขสินค้าต้องเป็นตัวเลข")
		return
	}
	var body struct {
		Rating  int    `json:"rating"`
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	comment := strings.TrimSpace(body.Comment)
	if body.Rating < 1 || body.Rating > 5 || comment == "" || utf8.RuneCountInString(comment) > 1000 {
		writeError(w, http.StatusBadRequest, "BAD_REVIEW", "ให้คะแนน 1 ถึง 5 และเขียนความคิดเห็นไม่เกิน 1000 ตัวอักษร")
		return
	}
	who := mw.WhoFrom(r.Context())
	err := h.Store.CreateReview(r.Context(), id, who.ID, body.Rating, comment)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "ไม่พบสินค้านี้")
		return
	}
	if errors.Is(err, store.ErrAlreadyReviewed) {
		writeError(w, http.StatusConflict, "ALREADY_REVIEWED", "คุณรีวิวสินค้านี้แล้ว")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true})
}

func (h *Handler) ListAllReviews(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListAllReviews(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขรีวิวต้องเป็นตัวเลข")
		return
	}
	err := h.Store.DeleteReview(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "REVIEW_NOT_FOUND", "ไม่พบรีวิวนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListWishlist(w http.ResponseWriter, r *http.Request) {
	who := mw.WhoFrom(r.Context())
	items, err := h.Store.ListWishlist(r.Context(), who.ID)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ToggleWishlist(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "productId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขสินค้าต้องเป็นตัวเลข")
		return
	}
	who := mw.WhoFrom(r.Context())
	liked, err := h.Store.ToggleWishlist(r.Context(), who.ID, id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "ไม่พบสินค้านี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"liked": liked})
}

func (h *Handler) ListCart(w http.ResponseWriter, r *http.Request) {
	who := mw.WhoFrom(r.Context())
	items, err := h.Store.ListCart(r.Context(), who.ID)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

type cartReq struct {
	ProductID int64  `json:"productId"`
	Size      string `json:"size"`
	Quantity  int    `json:"quantity"`
}

func readCart(w http.ResponseWriter, r *http.Request) (cartReq, bool) {
	var body cartReq
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return body, false
	}
	body.Size = strings.TrimSpace(body.Size)
	if body.ProductID <= 0 || body.Size == "" || utf8.RuneCountInString(body.Size) > 20 {
		writeError(w, http.StatusBadRequest, "BAD_CART", "ต้องระบุสินค้าและไซซ์")
		return body, false
	}
	return body, true
}

func (h *Handler) AddCart(w http.ResponseWriter, r *http.Request) {
	body, ok := readCart(w, r)
	if !ok {
		return
	}
	if body.Quantity < 1 {
		body.Quantity = 1
	}
	who := mw.WhoFrom(r.Context())
	err := h.Store.AddCart(r.Context(), who.ID, body.ProductID, body.Size, body.Quantity)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "ไม่พบสินค้านี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true})
}

func (h *Handler) SetCart(w http.ResponseWriter, r *http.Request) {
	body, ok := readCart(w, r)
	if !ok {
		return
	}
	who := mw.WhoFrom(r.Context())
	err := h.Store.SetCart(r.Context(), who.ID, body.ProductID, body.Size, body.Quantity)
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

func (h *Handler) ListMyOrders(w http.ResponseWriter, r *http.Request) {
	who := mw.WhoFrom(r.Context())
	items, err := h.Store.ListBuyerOrders(r.Context(), who.ID)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) Checkout(w http.ResponseWriter, r *http.Request) {
	who := mw.WhoFrom(r.Context())
	err := h.Store.Checkout(r.Context(), who.ID)
	if errors.Is(err, store.ErrEmptyCart) {
		writeError(w, http.StatusBadRequest, "EMPTY_CART", "ตะกร้าว่าง")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true})
}

func (h *Handler) ListSellerOrders(w http.ResponseWriter, r *http.Request) {
	who := mw.WhoFrom(r.Context())
	items, err := h.Store.ListSellerOrders(r.Context(), who.ID)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ShipOrder(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขคำสั่งซื้อต้องเป็นตัวเลข")
		return
	}
	var body struct {
		TrackingNo string `json:"trackingNo"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	tracking := strings.TrimSpace(body.TrackingNo)
	if tracking == "" || utf8.RuneCountInString(tracking) > 50 {
		writeError(w, http.StatusBadRequest, "BAD_TRACKING", "กรุณากรอกเลขพัสดุ")
		return
	}
	who := mw.WhoFrom(r.Context())
	err := h.Store.ShipOrder(r.Context(), id, who.ID, tracking)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "ORDER_NOT_FOUND", "ไม่พบคำสั่งซื้อนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateDispute(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขสินค้าต้องเป็นตัวเลข")
		return
	}
	var body struct {
		Reason string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	reason := strings.TrimSpace(body.Reason)
	if reason == "" || utf8.RuneCountInString(reason) > 1000 {
		writeError(w, http.StatusBadRequest, "BAD_DISPUTE", "กรุณาระบุเหตุผลไม่เกิน 1000 ตัวอักษร")
		return
	}
	who := mw.WhoFrom(r.Context())
	err := h.Store.CreateDispute(r.Context(), id, who.ID, reason)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "ไม่พบสินค้านี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true})
}

func (h *Handler) ListDisputes(w http.ResponseWriter, r *http.Request) {
	items, err := h.Store.ListDisputes(r.Context())
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) ResolveDispute(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขคำร้องต้องเป็นตัวเลข")
		return
	}
	err := h.Store.ResolveDispute(r.Context(), id)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "DISPUTE_NOT_FOUND", "ไม่พบคำร้องนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListMessages(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขสินค้าต้องเป็นตัวเลข")
		return
	}
	who := mw.WhoFrom(r.Context())
	items, err := h.Store.ListMessages(r.Context(), id, who.ID)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (h *Handler) CreateMessage(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขสินค้าต้องเป็นตัวเลข")
		return
	}
	var body struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}
	text := strings.TrimSpace(body.Text)
	if text == "" || utf8.RuneCountInString(text) > 500 {
		writeError(w, http.StatusBadRequest, "BAD_MESSAGE", "ข้อความว่าง หรือยาวเกิน 500 ตัวอักษร")
		return
	}
	who := mw.WhoFrom(r.Context())
	err := h.Store.AddMessage(r.Context(), id, who.ID, text)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "ไม่พบสินค้านี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"ok": true})
}

type productBody struct {
	Name     string `json:"name"`
	Category string `json:"category"`
	Price    int    `json:"price"`
	Stock    int    `json:"stock"`
	Image    string `json:"image"`
}

func readProduct(w http.ResponseWriter, r *http.Request) (store.ProductInput, bool) {
	var body productBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return store.ProductInput{}, false
	}
	in := store.ProductInput{
		Name:     strings.TrimSpace(body.Name),
		Category: strings.TrimSpace(body.Category),
		Price:    body.Price,
		Stock:    body.Stock,
		Image:    strings.TrimSpace(body.Image),
	}
	switch {
	case in.Name == "" || utf8.RuneCountInString(in.Name) > 255:
		writeError(w, http.StatusBadRequest, "BAD_PRODUCT", "ต้องมีชื่อสินค้า")
	case in.Category != "เสื้อ" && in.Category != "กางเกง" && in.Category != "เดรส":
		writeError(w, http.StatusBadRequest, "BAD_PRODUCT", "หมวดต้องเป็น เสื้อ กางเกง หรือ เดรส")
	case in.Price < 0 || in.Stock < 0:
		writeError(w, http.StatusBadRequest, "BAD_PRODUCT", "ราคาและสต็อกต้องไม่ติดลบ")
	default:
		return in, true
	}
	return store.ProductInput{}, false
}

func (h *Handler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	in, ok := readProduct(w, r)
	if !ok {
		return
	}
	who := mw.WhoFrom(r.Context())
	item, err := h.Store.CreateProduct(r.Context(), who.ID, in)
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	id, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขสินค้าต้องเป็นตัวเลข")
		return
	}
	in, ok := readProduct(w, r)
	if !ok {
		return
	}
	who := mw.WhoFrom(r.Context())
	item, err := h.Store.UpdateProduct(r.Context(), id, who.ID, in)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "PRODUCT_NOT_FOUND", "ไม่พบสินค้านี้ หรือไม่ใช่ของร้านคุณ")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}
