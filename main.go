package eskizuz

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/realtemirov/logt"
)

const (
	baseURL    string = "https://notify.eskiz.uz/api"
	smsURL     string = "/message/sms/send"
	userURL    string = "/auth/user"
	loginURL   string = "/auth/login"
	refreshURL string = "/auth/refresh"

	limitURL string = "/user/get-limit"
)

type Auth struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Eskiz struct {
	Token   string
	Type    string
	log     logt.ILog
	Message string
	Error   string
}
type data struct {
	err         error
	message     string
	code        int
	dataMarshal string
}

type SMS struct {
	MobilePhone string `json:"mobile_phone"`
	Message     string `json:"message"`
	From        string `json:"from"`
	CallbackURL string `json:"callback_url"`
}

// Send sms
//
// Authorization: Bearer token
//
//	sms := SMS{
//		MobilePhone:  "998771234567",
//		Message:      "test-message",
//		From:         "go-eskiz-uz",
//		CallbackURL:  "https://eskiz.uz",
//	}
//	eskiz.Send(&sms)
func (eskiz *Eskiz) Send(sms *SMS) (map[string]interface{}, error) {

	w := eskiz.log.NewWriter("send sms", false)
	defer w.Close()

	key, value := headerAuth(eskiz.Token, eskiz.Type)
	req := request(w, baseURL+smsURL, "POST", sms, key, value)
	if e(w, req.err, "in send sms") {
		return nil, req.err
	}

	response := map[string]interface{}{}
	err := json.Unmarshal([]byte(req.dataMarshal), &response)
	if e(w, err, "in send sms") {
		return nil, err
	}

	return response, nil
}

// Get user sms-limit
//
// Authorization: Bearer token
func (eskiz *Eskiz) GetUserLimit() (map[string]interface{}, error) {

	w := eskiz.log.NewWriter("get user limit", false)
	defer w.Close()

	key, value := headerAuth(eskiz.Token, eskiz.Type)
	req := request(w, baseURL+limitURL, "GET", nil, key, value)
	if e(w, req.err, "in get user limit") {
		return nil, req.err
	}

	response := map[string]interface{}{}
	err := json.Unmarshal([]byte(req.dataMarshal), &response)
	if e(w, err, "in get user limit") {
		return nil, err
	}

	return response, nil
}

// Get user info
//
// Authorization: Bearer token
func (eskiz *Eskiz) GetMe() (map[string]interface{}, error) {

	w := eskiz.log.NewWriter("get me", false)
	defer w.Close()

	key, value := headerAuth(eskiz.Token, eskiz.Type)
	req := request(w, baseURL+userURL, "GET", nil, key, value)
	if e(w, req.err, "in get me") {
		return nil, req.err
	}

	response := map[string]interface{}{}
	err := json.Unmarshal([]byte(req.dataMarshal), &response)
	if e(w, err, "in get me") {
		return nil, err
	}

	return response, nil
}

// Refresh token
//
// Authorization: Bearer token
func (eskiz *Eskiz) RefreshToken() error {

	w := eskiz.log.NewWriter("refresh token", false)
	defer w.Close()

	key, value := headerAuth(eskiz.Token, eskiz.Type)

	req := request(w, baseURL+refreshURL, "PATCH", nil, key, value)
	if e(w, req.err, "in refresh token") {
		return req.err
	}

	response := map[string]interface{}{}

	err := json.Unmarshal([]byte(req.dataMarshal), &response)
	if e(w, err, "in refresh token") {
		return err
	}

	eskiz.Token = response["data"].(map[string]interface{})["token"].(string)
	eskiz.Type = response["token_type"].(string)
	eskiz.Message = response["message"].(string)

	return nil
}

// Login and get token
//
//	auth := Auth{
//			Email:    "your_email",
//			Password: "your_sms_service_password",
//	}
//
//	eskiz, err := GetToken(&auth)
//	if err != nil {
//		panic(err)
//	}
func GetToken(auth *Auth) (*Eskiz, error) {

	l := logt.NewLog(&logt.Log{
		NameSpace: "eskiz",
	})

	w := l.NewWriter("authorization", false)
	defer w.Close()

	req := request(w, baseURL+loginURL, "POST", auth, "", "")
	if e(w, req.err, "in authorization") {
		return &Eskiz{
			Token:   "",
			Type:    "",
			log:     l,
			Message: req.message,
			Error:   req.err.Error(),
		}, req.err
	}

	response := map[string]interface{}{}

	err := json.Unmarshal([]byte(req.dataMarshal), &response)
	if e(w, err, "in authorization") {
		return &Eskiz{
			Token:   "",
			Type:    "",
			log:     l,
			Message: req.message,
			Error:   err.Error(),
		}, err
	}

	return &Eskiz{
		Token:   response["data"].(map[string]interface{})["token"].(string),
		Type:    response["token_type"].(string),
		log:     l,
		Message: response["message"].(string),
		Error:   "",
	}, nil
}

func request(w logt.IWriter, url string, method string, body interface{}, keyHeader, valueHeader string) data {

	js, err := json.Marshal(&body)
	if e(w, err, "marshaling body") {
		return data{
			err:         err,
			message:     "marshaling body",
			code:        -1,
			dataMarshal: "",
		}
	}
	client := &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	req, err := http.NewRequest(method, url, bytes.NewBuffer(js))
	if err != nil {
		w.Error("init new request body", url, body, err)
		return data{
			err:         err,
			message:     "new request",
			dataMarshal: "",
			code:        -1,
		}
	}

	if keyHeader != "" {

		req.Header.Set(keyHeader, valueHeader)
	}

	// Create a new context with the desired timeout
	ctx, cancel := context.WithTimeout(req.Context(), 60*time.Second)
	defer cancel()

	// Associate the context with the request
	req = req.WithContext(ctx)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := client.Do(req)
	if err != nil {
		w.Error("doing request", url, body, err)
		return data{
			err:         err,
			message:     "doing request",
			dataMarshal: "",
			code:        -1,
		}
	}
	defer resp.Body.Close()

	str := strings.Builder{}

	// body to string
	buf := new(bytes.Buffer)
	buf.ReadFrom(resp.Body)
	str.WriteString(buf.String())
	out := map[string]interface{}{}

	json.Unmarshal([]byte(str.String()), &out)

	w.Succes("response", url, body, out)

	switch resp.StatusCode {

	case 400:
		return data{
			err:         fmt.Errorf("bad request"),
			message:     "bad request",
			dataMarshal: str.String(),
			code:        400,
		}
	case 401:
		return data{
			err:         fmt.Errorf("unauthorized"),
			message:     "unauthorized",
			dataMarshal: str.String(),
			code:        401,
		}
	}

	return data{
		err:         nil,
		message:     "succes",
		code:        200,
		dataMarshal: str.String(),
	}
}

func e(w logt.IWriter, err error, message string) bool {
	if err != nil {
		w.Error(message, err)
		return true
	}
	return false
}

func headerAuth(token, _type string) (string, string) {
	return "Authorization", strings.ToUpper(string(_type[0])) + _type[1:] + " " + token
}
