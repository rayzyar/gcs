package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"net/http/httptest"

	redis "github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
	geo "github.com/kellydunn/golang-geo"
	"github.com/rayzyar/gcs/gcs-backend/dto"
	"github.com/rayzyar/gcs/gcs-backend/prebookdata"
	"github.com/rayzyar/gcs/pkg/email"
	"github.com/rayzyar/gcs/pkg/redis"
	"github.com/rayzyar/gcs/pkg/redisgeo"
)

const (
	logTag          = "gcs"
	resthost        = "172.16.2.114:8080"
	distInfinity    = math.MaxInt64
	KeybookingCount = "booking:count"
	nearRadius      = 3
)

var center = geo.NewPoint(1.2966426, 103.7742052)

func init() {
	for _, d := range dto.RegisteredReceiver {
		recorder := httptest.NewRecorder()
		handleReceiveRegisterDto(recorder, d)
	}
}

func main() {
	r := mux.NewRouter()
	// wire up CatchAll
	r.NotFoundHandler = http.NotFoundHandler()
	r.HandleFunc("/give", giveHandler).Methods("POST")
	r.HandleFunc("/give/current", giveCurrentHandler).Methods("GET")
	r.HandleFunc("/receive/register", receiveRegisterHandler).Methods("POST")
	r.HandleFunc("/receive/list", receiveListHandler).Methods("GET")
	r.HandleFunc("/receive/confirm", receiveConfirmHandler).Methods("GET")
	srv := &http.Server{
		Handler: r,
		Addr:    resthost,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Println(logTag, "rest server started")
	log.Fatal(srv.ListenAndServe())
}

func giveHandler(resp http.ResponseWriter, req *http.Request) {
	var body []byte
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read request err:%s", err)
		errResp(resp, msg)
		return
	}
	defer req.Body.Close()
	giveReq := &dto.GiveRequest{}
	if err = json.Unmarshal(body, giveReq); err != nil {
		msg := fmt.Sprintf("failed to read request err:%s", err)
		errResp(resp, msg)
		return
	}
	fmt.Println("req:", giveReq)
	pbCode, err := newPreBookCode()
	if err != nil {
		msg := fmt.Sprintf("failed to generate prebook code err:%s", err)
		errResp(resp, msg)
	}

	bdata := &prebookdata.PrebookData{
		PreBookCode: pbCode,
		Item:        giveReq.Item,
		CityID:      giveReq.CityID,
		State:       prebookdata.StateAllocating,
		WeightKG:    giveReq.WeightKG,
		PickUp:      geo.NewPoint(giveReq.Lat, giveReq.Lng),
	}
	prebookdata.DaoV1.StorePrebookData(bdata)

	outputData, err := json.Marshal(bdata)
	if err != nil {
		msg := fmt.Sprintf("failed to generate prebook code err:%s", err)
		errResp(resp, msg)
	}

	go func() {
		err := allocate(bdata)
		if err != nil {
			fmt.Println("allocation failed err:", err)
		}
	}()

	fmt.Println(string(outputData))
	resp.WriteHeader(http.StatusCreated)
	resp.Write(outputData)
	return

}

func allocate(bdata *prebookdata.PrebookData) error {
	loc := findNearestDemand(bdata)
	if loc == nil {
		return errors.New("not matched")
	}
	fmt.Printf("filtered loc:%#v\n", loc)
	bdata.State = prebookdata.StateAllocated
	bdata.DriverName = "Driver A"
	bdata.DriverPhoneNumber = "+6593004400"
	bdata.PlateNumber = ""
	bdata.PickUpTime = time.Now().Add(time.Minute).UnixNano() / 1000
	bdata.DemandID, _ = strconv.ParseInt(loc.Name[4:], 10, 64)
	// should have been retrieved from DemandID
	destEmail := "ray.zezhou@gmail.com"
	email.Send(destEmail, "http://"+resthost+"/confirm/\n\n"+
		"phone number: "+bdata.DriverPhoneNumber+"\n\n"+
		"courier plate: "+bdata.PlateNumber+"\n\n"+
		"pick up time: "+time.Unix(bdata.PickUpTime/int64(time.Millisecond), 0).Format(time.RFC822)+"\n\n"+
		"driver: "+bdata.DriverName,
	)
	return nil
}

func findNearestDemand(bdata *prebookdata.PrebookData) *redisgeo.GeoLocation {
	conn := rediscli.GetConn()
	defer conn.Close()
	locations, _ := redisgeo.GeoLocations(conn.Do("GEORADIUS", cityKey(bdata.CityID), bdata.PickUp.Lng(), bdata.PickUp.Lat(), nearRadius, "km", "WITHCOORD", "WITHDIST", "WITHHASH"))
	loc := filter(conn, bdata.WeightKG, locations)
	return loc
}

func filter(conn redis.Conn, weight float64, locations []*redisgeo.GeoLocation) *redisgeo.GeoLocation {
	fmt.Printf("%d locations selected\n", len(locations))
	for _, loc := range locations {
		b, _ := redis.Bytes(conn.Do("GET", loc.Name))
		var buffer = &bytes.Buffer{}
		_, _ = buffer.Write(b)
		var rcvRegReq = &dto.ReceiveRegisterRequest{}
		dec := gob.NewDecoder(buffer)
		dec.Decode(rcvRegReq)
		fmt.Printf("demand weight:%f, requested weight:%f\n", rcvRegReq.Demand, weight)
		if weight > rcvRegReq.Demand*0.6 && weight < rcvRegReq.Demand*1.5 {
			fmt.Printf("loc:%#v, allocated\n", loc)
			return loc
		}
		fmt.Printf("loc:%#v, filtered out\n", loc)
	}
	return nil
}

func giveCurrentHandler(resp http.ResponseWriter, req *http.Request) {
	values, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		errResp(resp, "failed to parse url raw query err:%s", err)
		return
	}

	preBookCode := values.Get("preBookCode")
	bdata, err := prebookdata.DaoV1.Get(preBookCode)
	if err != nil {
		errResp(resp, "failed to get prebookdata err:%s", err)
	}
	outputData, err := json.Marshal(bdata)
	if err != nil {
		msg := fmt.Sprintf("failed to generate prebook code err:%s", err)
		errResp(resp, msg)
	}
	fmt.Println(string(outputData))
	resp.WriteHeader(http.StatusOK)
	resp.Write(outputData)
}

func handleReceiveRegisterDto(resp http.ResponseWriter, rcvRegReq *dto.ReceiveRegisterRequest) {
	conn := rediscli.GetConn()
	defer conn.Close()
	_, err := conn.Do("GEOADD", cityKey(rcvRegReq.CityID), rcvRegReq.Lng, rcvRegReq.Lat, memberKey(rcvRegReq.UserID))
	if err != nil {
		msg := fmt.Sprintf("failed to add geo point err:%s", err)
		errResp(resp, msg)
		return
	}

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	err = enc.Encode(rcvRegReq)
	_, err = conn.Do("SET", memberKey(rcvRegReq.UserID), b.String())
	if err != nil {
		msg := fmt.Sprintf("failed to add geo point err:%s", err)
		errResp(resp, msg)
		return
	}
	resp.WriteHeader(http.StatusOK)
	return
}

func receiveRegisterHandler(resp http.ResponseWriter, req *http.Request) {
	var body []byte
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		errResp(resp, "failed to read request err:%s", err)
		return
	}
	defer req.Body.Close()
	rcvRegReq := &dto.ReceiveRegisterRequest{}
	if err = json.Unmarshal(body, rcvRegReq); err != nil {
		errResp(resp, "failed to read request err:%s", err)
		return
	}
	handleReceiveRegisterDto(resp, rcvRegReq)
}

func receiveListHandler(resp http.ResponseWriter, req *http.Request) {
	values, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		errResp(resp, "failed to parse url raw query err:%s", err)
		return
	}

	cityID, err := strconv.ParseInt(values.Get("cityID"), 10, 64)
	if err != nil {
		errResp(resp, "failed to parse cityID err:%s", err)
		return
	}

	conn := rediscli.GetConn()
	defer conn.Close()
	locations, err := redisgeo.GeoLocations(conn.Do("GEORADIUS", cityKey(cityID), center.Lng(), center.Lat(), distInfinity, "km", "WITHCOORD", "WITHDIST", "WITHHASH"))
	if err != nil {
		errResp(resp, "failed to get position from redis err:%s", err)
		return
	}

	output, err := json.Marshal(locations)
	if err != nil {
		errResp(resp, "failed to marshal locations err:%s", err)
		return
	}
	resp.WriteHeader(http.StatusOK)
	resp.Write(output)
	return
}

func receiveConfirmHandler(resp http.ResponseWriter, req *http.Request) {
	values, err := url.ParseQuery(req.URL.RawQuery)
	if err != nil {
		errResp(resp, "failed to parse url raw query err:%s", err)
		return
	}
	token := values.Get("token")
	pCode := prebookdata.DaoV1.GetByToken(token)
	if pCode == "" {
		errResp(resp, "token is invalid")
		return
	}

}
func errResp(resp http.ResponseWriter, msg string, args ...interface{}) {
	resp.WriteHeader(http.StatusInternalServerError)
	resp.Write([]byte(fmt.Sprintf(msg, args...)))
}

func cityKey(cityID int64) string {
	return "city:" + strconv.FormatInt(cityID, 10)
}

func memberKey(userID int64) string {
	return "user:" + strconv.FormatInt(userID, 10)
}

func newPreBookCode() (string, error) {
	conn := rediscli.GetConn()
	defer conn.Close()
	count, err := redis.Int64(conn.Do("INCR", KeybookingCount))
	if err != nil {
		return "", err
	}
	return "ABC-" + strconv.FormatInt(count, 10), nil
}
