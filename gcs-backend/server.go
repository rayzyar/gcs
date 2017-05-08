package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
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
	"github.com/rayzyar/gcs/pkg/redis"
	"github.com/rayzyar/gcs/pkg/redisgeo"
)

const (
	logTag          = "gcs"
	resthost        = "172.16.2.114:8080"
	distInfinity    = math.MaxInt64
	KeybookingCount = "booking:count"
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

	bdata := prebookdata.PrebookData{
		PreBookCode: pbCode,
		State:       prebookdata.StateAllocating,
		WeightKG:    giveReq.WeightKG,
	}

	outputData, err := json.Marshal(bdata)
	if err != nil {
		msg := fmt.Sprintf("failed to generate prebook code err:%s", err)
		errResp(resp, msg)
	}
	fmt.Println(string(outputData))
	resp.WriteHeader(http.StatusCreated)
	resp.Write(outputData)
	return

}

func giveCurrentHandler(resp http.ResponseWriter, req *http.Request) {
	var body []byte
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		msg := fmt.Sprintf("failed to read request err:%s", err)
		errResp(resp, msg)
		return
	}
	defer req.Body.Close()
	giveCurrentReq := &dto.GiveCurrentRequest{}
	if err = json.Unmarshal(body, giveCurrentReq); err != nil {
		msg := fmt.Sprintf("failed to read request err:%s", err)
		errResp(resp, msg)
		return
	}
	fmt.Println("req:", giveCurrentReq)
	bdata := prebookdata.DaoV1.Get(giveCurrentReq.PreBookCode)
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
	_, err := conn.Do("GEOADD", cityKey(rcvRegReq.CityID), rcvRegReq.Lat, rcvRegReq.Lng, memberKey(rcvRegReq.UserID))
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
