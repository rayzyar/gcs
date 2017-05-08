package redisgeo

import (
	"errors"

	redis "github.com/garyburd/redigo/redis"
)

// GeoPoint Represents a Physical GeoPoint in geographic notation [lat, lng].
type GeoPoint struct {
	lat float64
	lng float64
}

// NewGeoPoint Returns a new GeoPoint populated by the passed in latitude (lat) and longitude (lng) values.
func NewGeoPoint(lat float64, lng float64) GeoPoint {
	return GeoPoint{lat: lat, lng: lng}
}

// Lat Returns GeoPoint p's latitude.
func (p GeoPoint) Lat() float64 {
	return p.lat
}

// Lng Returns GeoPoint p's longitude.
func (p GeoPoint) Lng() float64 {
	return p.lng
}

// GeoLocation is a struct which represents the response of some geo related commands
type GeoLocation struct {
	Name     string
	Distance float64
	GeoPoint GeoPoint
	Hash     int64
}

// GeoPosition converts a reply to the GeoPoint
// For the response of the below commands
// https://redis.io/commands/geopos
func GeoPosition(reply interface{}, err error) ([]GeoPoint, error) {
	values, err := redis.Values(reply, err)
	if err != nil {
		return nil, err
	}
	geoPoints := make([]GeoPoint, len(values))
	for i, value := range values {
		geoPoints[i], err = convertToGeoGeoPoint(value)
		if err != nil {
			return nil, err
		}
	}
	return geoPoints, nil
}

// GeoLocations converts a reply to the []*GeoLocation
// For the response of the below commands
// https://redis.io/commands/georadius
// https://redis.io/commands/georadiusbymember
func GeoLocations(reply interface{}, err error) ([]*GeoLocation, error) {
	values, err := redis.Values(reply, err)
	if err != nil {
		return nil, err
	}
	list := make([]*GeoLocation, len(values))
	for i, v := range values {
		list[i], err = convertToGeoLocation(v)
		if err != nil {
			return nil, err
		}
	}
	return list, nil
}

// convertToGeoGeoPoint converts interface{} to GeoPoint
func convertToGeoGeoPoint(reply interface{}) (GeoPoint, error) {
	latlng, err := Float64s(reply, nil)
	if err != nil {
		return GeoPoint{}, err
	}
	if len(latlng) != 2 {
		return GeoPoint{}, errors.New("Err Convert To GeoPoint with Invalid Length")
	}
	return NewGeoPoint(latlng[1], latlng[0]), nil
}

// convertToGeoLocation converts []interface{} to a GeoLocation struct
func convertToGeoLocation(reply interface{}) (*GeoLocation, error) {
	result := &GeoLocation{}
	pReply, err := redis.Values(reply, nil)
	if err != nil {
		return nil, err
	}
	pReplyLen := len(pReply)
	if pReplyLen == 0 || pReplyLen > 4 {
		return nil, errors.New("Err Convert To GeoLocation with Invalid Length")
	}
	for i, item := range pReply {
		if i == 0 {
			result.Name, err = redis.String(item, nil)
			if err != nil {
				return nil, err
			}
			continue
		}
		switch item.(type) {
		case int64:
			result.Hash, err = redis.Int64(item, nil)
			if err != nil {
				return nil, err
			}
		case []byte:
			result.Distance, err = redis.Float64(item, nil)
			if err != nil {
				return nil, err
			}
		case []interface{}:
			result.GeoPoint, err = convertToGeoGeoPoint(item)
			if err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

// Float64s is a helper that converts an array command reply to a []Float64. If
// err is not equal to nil, then Float64s returns nil, err.
func Float64s(reply interface{}, err error) ([]float64, error) {
	values, err := redis.Values(reply, err)
	if err != nil {
		return nil, err
	}
	floats := make([]float64, len(values))
	for i, value := range values {
		f, err := redis.Float64(value, nil)
		if err != nil {
			return nil, err
		}
		floats[i] = f
	}
	return floats, nil
}
