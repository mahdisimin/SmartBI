package SQLServer

import (
	"fmt"
	external "intelligentBI/external/synops/ver1"
	"intelligentBI/pkg"
	"testing"
	"time"
)

func TestSynops_LoginHostory(t *testing.T) {
	db := requireDB(t)
	synopsServ := Synops{DB: db, resourseName: pkg.LoginHistory}

	// A user agent unique to this run tags the rows it inserts, so cleanup
	// removes exactly those and the test is re-runnable.
	userAgent := fmt.Sprintf("intelligentbi-test/%d", time.Now().UnixNano())
	t.Cleanup(func() { db.Exec("DELETE FROM synops.LoginHistory WHERE UserAgent = @p1", userAgent) })

	var resBodydata []external.LoginHistoryResBodydata
	resBodydata = []external.LoginHistoryResBodydata{
		external.LoginHistoryResBodydata{
			IpAddress: "127.0.0.1",
			Country:   "unknown",
			UserAgent: userAgent,
			TimeStamp: time.Now(),
			IsSuccess: false,
		}, external.LoginHistoryResBodydata{
			IpAddress: "127.0.0.2",
			Country:   "unknown",
			UserAgent: userAgent,
			TimeStamp: time.Now(),
			IsSuccess: false,
		}, external.LoginHistoryResBodydata{
			IpAddress: "127.0.0.3",
			Country:   "unknown",
			UserAgent: userAgent,
			TimeStamp: time.Now(),
			IsSuccess: false,
		}, external.LoginHistoryResBodydata{
			IpAddress: "127.0.0.4",
			Country:   "unknown",
			UserAgent: userAgent,
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
