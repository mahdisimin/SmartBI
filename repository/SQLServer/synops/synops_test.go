package SQLServer

import (
	external "intelligentBI/external/synops/ver1"
	"intelligentBI/pkg"
	"testing"
	"time"
)

func TestSynops_LoginHostory(t *testing.T) {
	synopsServ := Synops{resourseName: pkg.LoginHistory}

	var resBodydata []external.LoginHistoryResBodydata
	resBodydata = []external.LoginHistoryResBodydata{
		external.LoginHistoryResBodydata{
			IpAddress: "127.0.0.1",
			Country:   "unknown",
			UserAgent: "PostmanRuntime/7.48.0",
			TimeStamp: time.Now(),
			IsSuccess: false,
		}, external.LoginHistoryResBodydata{
			IpAddress: "127.0.0.2",
			Country:   "unknown",
			UserAgent: "PostmanRuntime/7.48.0",
			TimeStamp: time.Now(),
			IsSuccess: false,
		}, external.LoginHistoryResBodydata{
			IpAddress: "127.0.0.3",
			Country:   "unknown",
			UserAgent: "PostmanRuntime/7.48.0",
			TimeStamp: time.Now(),
			IsSuccess: false,
		}, external.LoginHistoryResBodydata{
			IpAddress: "127.0.0.4",
			Country:   "unknown",
			UserAgent: "PostmanRuntime/7.48.0",
			TimeStamp: time.Now(),
			IsSuccess: false,
		},
	}

	data := external.LoginHistoryRes{
		Status:  "OK",
		Code:    200,
		Message: "جزئیات تاریخچه با موفقیت ایجاد گردید.",
		Body: external.LoginHistoryResBody{
			Total:       3,
			CurrentPage: 1,
			Data:        resBodydata,
		},
	}

	if err := synopsServ.LoginHistory(data); err != nil {
		t.Error(err)
	}

}
