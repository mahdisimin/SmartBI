package ver1

import (
	"encoding/json"
	"errors"
	"intelligentBI/pkg"
	"io"
	"net/http"
	"time"
)

type Synops struct {
	endpoint pkg.SynOpsAPIList
}

type LoginHistoryResBodydata struct {
	IpAddress string    `json:"ip_address"`
	Country   string    `json:"country"`
	UserAgent string    `json:"user_agent"`
	TimeStamp time.Time `json:"time_stamp"`
	IsSuccess bool      `json:"is_success"`
}

type LoginHistoryResBody struct {
	Total       int8                      `json:"total"`
	CurrentPage int8                      `json:"current_page"`
	Data        []LoginHistoryResBodydata `json:"data"`
}
type LoginHistoryRes struct {
	Status  string              `json:"status"`
	Code    int                 `json:"code"`
	Message string              `json:"message"`
	Body    LoginHistoryResBody `json:"body"`
}

func (s Synops) Request() (any, error) {
	switch s.endpoint {
	case pkg.LoginHistory:
		return s.LoginHistory()

	}
	return nil, errors.New("unknown endpoint")
}

func (s Synops) HealthCheck() int8 {

	return 0
}

func (s Synops) LoginHistory() (LoginHistoryRes, error) {
	var response LoginHistoryRes
	req, _ := http.NewRequest(http.MethodGet, string(pkg.Req_LoginHostory), nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()
	bodybyte, _ := io.ReadAll(resp.Body)
	if err := json.Unmarshal(bodybyte, &response); err != nil {
		return response, err
	}
	return response, nil
}
