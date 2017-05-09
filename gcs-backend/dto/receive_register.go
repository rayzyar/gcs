package dto

import "time"

type ReceiveRegisterRequest struct {
	CityID     int64   `json:"cityID"`
	UserID     int64   `json:"userID"`
	Lng        float64 `json:"longitude"`
	Lat        float64 `json:"latitude"`
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
			Lng:    103.85334,
			Lat:    1.29396,
			Demand: 20,
			TimeRanges: []TimeRange{
				{StartTime: 0, EndTime: 24 * time.Hour},
			},
		},
		{
			CityID: 1,
			UserID: 2,
			Lng:    103.85106,
			Lat:    1.2973,
			Demand: 40,
			TimeRanges: []TimeRange{
				{StartTime: 0, EndTime: 24 * time.Hour},
			},
		},
		{
			CityID: 1,
			UserID: 3,
			Lng:    103.84903,
			Lat:    1.29436,
			Demand: 20,
			TimeRanges: []TimeRange{
				{StartTime: 0, EndTime: 24 * time.Hour},
			},
		},
		{
			CityID: 1,
			UserID: 4,
			Lng:    103.82387,
			Lat:    1.30483,
			Demand: 30,
			TimeRanges: []TimeRange{
				{StartTime: 0, EndTime: 24 * time.Hour},
			},
		},
		{
			CityID: 1,
			UserID: 5,
			Lng:    103.90547,
			Lat:    1.32094,
			Demand: 10,
			TimeRanges: []TimeRange{
				{StartTime: 0, EndTime: 24 * time.Hour},
			},
		},
		{
			CityID: 1,
			UserID: 6,
			Lng:    103.91009,
			Lat:    1.32064,
			Demand: 100,
			TimeRanges: []TimeRange{
				{StartTime: 0, EndTime: 24 * time.Hour},
			},
		},
	}
)
