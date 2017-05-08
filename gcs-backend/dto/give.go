package dto

type GiveRequest struct {
	CityID     int64   `json:"cityID"`
	UserID     int64   `json:"userID"`
	Lat        float64 `json:"latitude"`
	Lng        float64 `json:"longitude"`
	WeightKG   float64 `json:"weight"`
	ExpireTime int64   `json:"expire"`
}
