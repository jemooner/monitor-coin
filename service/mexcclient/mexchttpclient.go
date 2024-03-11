package mexcclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"monitor-coin/commonlib"
	"monitor-coin/commonlib/dlog"
	"net/http"
	"strings"
	"time"
)

var (
	httpClient *http.Client
	uriList    map[string]string
	callee     = "mexc"
)

func InitMexcClient(cfg *commonlib.HttpClientConf) {
	if uriList == nil || len(uriList) == 0 {
		uriList = make(map[string]string)
		uriList["GetAllCoins"] = "/api/v3/exchangeInfo"
		uriList["Get24hrPriceTicker"] = "/api/v3/ticker/24hr"
	}
	httpClient = commonlib.GetDefaultHttpClient(cfg.TimeoutSec)
	fmt.Println("InitMexcClient done")
}

func getEndPoint(uri string) string {
	uri, ok := uriList[uri]
	if ok && len(uri) > 0 {
		return commonlib.LaunchConfig().Mexc.Host[0] + uri
	}
	return ""
}

func doHttpRequest(ctx context.Context, m string, ep string, ps map[string]interface{}) (r []byte, err error) {
	ts, traceId := time.Now(), commonlib.GetTrace(ctx)
	dlog.Infof(commonlib.CallInLog(ctx, callee, m, ep, ps))

	var request *http.Request
	if m == http.MethodGet {
		request, err = http.NewRequest(http.MethodGet, ep, nil)
		if err != nil {
			return nil, err
		}
		if ps != nil && len(ps) > 0 {
			qs := request.URL.Query()
			for k, v := range ps {
				qs.Add(k, fmt.Sprint(v))
			}
			request.URL.RawQuery = qs.Encode()
		}
	} else if m == http.MethodPost {
		var psb []byte
		if ps != nil && len(ps) > 0 {
			psb, err = json.Marshal(ps)
			if err != nil {
				return nil, err
			}
		}
		request, err = http.NewRequest(http.MethodPost, ep, strings.NewReader(string(psb)))
	}
	if err != nil {
		dlog.Errorf("%v||doHttpRequest NewRequest fail, err=%v", traceId, err)
		return nil, err
	}

	resp, err := httpClient.Do(request)
	dlog.Infof(commonlib.CallOutLog(ctx, callee, []byte(""), err, ts))

	if err != nil || resp == nil || resp.Body == nil {
		err = fmt.Errorf("doHttpRequest fail,err=%v", err)
		dlog.Errorf("%v||err=%v", traceId, err)
		return
	}
	defer resp.Body.Close()

	r, err = ioutil.ReadAll(resp.Body)
	if err != nil || len(r) == 0 || resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("doHttpRequest read resp.body fail: %v", err)
	}
	ss := r
	if len(ss) > 1500 {
		ss = ss[0:1500]
	}
	dlog.Infof("%v||ReadAll=%s,err=%v", traceId, ss, err)

	return
}
