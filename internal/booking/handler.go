package booking

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/B8rislav/cinema-booking/internal/utils"
)

type handler struct {
	svc *Service
}

type seatInfo struct {
	SeatID    string `json:"seat_id"`
	UserID    string `json:"user_id"`
	Booked    bool   `json:"booked"`
	Confirmed bool   `json:"confirmed"`
}

func NewHandler(svc *Service) *handler {
	return &handler{svc}
}

func (h *handler) ListSeats(w http.ResponseWriter, r *http.Request) {
	movieID := r.PathValue("movieID")

	bookings := h.svc.store.ListBookings(movieID)

	seats := make([]seatInfo, 0, len(bookings))
	for _, b := range bookings {
		log.Println(b.Status)
		seats = append(seats, seatInfo{
			SeatID:    b.SeatID,
			UserID:    b.UserID,
			Booked:    b.Status == "held",
			Confirmed: b.Status == "confirmed",
		})
	}

	utils.WriteJSON(w, http.StatusOK, seats)
}

func (h *handler) HoldSeat(w http.ResponseWriter, r *http.Request) {
	movieID := r.PathValue("movieID")
	seatID := r.PathValue("seatID")

	type holdRequest struct {
		UserID string `json:"user_id"`
	}
	var req holdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println(err)
		return
	}
	data := Booking{
		MovieID: movieID,
		SeatID:  seatID,
		UserID:  req.UserID,
	}

	session, err := h.svc.Book(data)
	fmt.Println(session)
	if err != nil {
		log.Println(err)
		return
	}

	type holdResponse struct {
		SessionID string `json:"session_id"`
		MovieID   string `json:"movie_id"`
		SeatID    string `json:"seat_id"`
		ExpiresAt string `json:"expires_at"`
	}

	utils.WriteJSON(w, http.StatusCreated, holdResponse{
		SeatID:    seatID,
		MovieID:   session.MovieID,
		SessionID: session.ID,
		ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
	})
}

func (h *handler) ConfirmSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")
	err := h.svc.store.ConfirmSession(sessionID)
	if err != nil {
		log.Println(err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, sessionID)

}

func (h *handler) ReleaseSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")
	err := h.svc.store.ReleaseSession(sessionID)
	if err != nil {
		log.Println(err)
		return
	}
	utils.WriteJSON(w, http.StatusOK, sessionID)
}
