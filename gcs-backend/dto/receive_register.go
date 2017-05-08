package dto

import "time"

type ReceiveRegisterRequest struct {
	CityID     int64   `json:"cityID"`
	UserID     int64   `json:"userID"`
	Lat        float64 `json:"latitude"`
	Lng        float64 `json:"longitude"`
	Demand     float64 `json:"demand"` // amount of food, unit: one set of meal for one adult
	TimeRanges []TimeRange
}

type TimeRange struct {
	StartTime time.Duration
	EndTime   time.Duration
}

var (
	RegisteredReceiver = []*ReceiveRegisterRequest{
		{
			CityID: 1,
			UserID: 1,
			Lng:    1.29396,
			Lat:    103.85334,
			Demand: 20,
			TimeRanges: []TimeRange{
				{StartTime: 0, EndTime: 24 * time.Hour},
			},
		},
		{
			CityID: 1,
			UserID: 2,
			Lng:    1.2973,
			Lat:    103.85106,
			Demand: 40,
			TimeRanges: []TimeRange{
				{StartTime: 0, EndTime: 24 * time.Hour},
			},
		},
		{
			CityID: 1,
			UserID: 3,
			Lng:    1.29436,
			Lat:    103.84903,
			Demand: 20,
			TimeRanges: []TimeRange{
				{StartTime: 0, EndTime: 24 * time.Hour},
			},
		},
		{
			CityID: 1,
			UserID: 4,
			Lng:    1.30483,
			Lat:    103.82387,
			Demand: 30,
			TimeRanges: []TimeRange{
				{StartTime: 0, EndTime: 24 * time.Hour},
			},
		},
		{
			CityID: 1,
			UserID: 5,
			Lng:    1.32094,
			Lat:    103.90547,
			Demand: 10,
			TimeRanges: []TimeRange{
				{StartTime: 0, EndTime: 24 * time.Hour},
			},
		},
		{
			CityID: 1,
			UserID: 6,
			Lng:    1.32064,
			Lat:    103.91009,
			Demand: 100,
			TimeRanges: []TimeRange{
				{StartTime: 0, EndTime: 24 * time.Hour},
			},
		},
	}
)
