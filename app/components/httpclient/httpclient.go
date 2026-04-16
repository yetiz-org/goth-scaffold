package httpclient

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/yetiz-org/gone/ghttp"
	buf "github.com/yetiz-org/goth-bytebuf"
	kklogger "github.com/yetiz-org/goth-kklogger"
)

const (
	maxLogBodyBytes        = 204800
	maxLogRequestBodyBytes = 200000
	maxLogBodyTrimBytes    = 102400

	// DefaultTimeout is the timeout applied to every outgoing HTTP request.
	// It covers dial, TLS handshake, wait for first byte, and full response read.
	DefaultTimeout = 30 * time.Second
)

// _client is the shared HTTP client used by Do and DoAndLog.
// Tests in this package may replace it temporarily to assert timeout behaviour.
var _client = &http.Client{Timeout: DefaultTimeout}

func NewRequest(method string, url string, body io.Reader) *http.Request {
	if req, err := http.NewRequest(method, url, body); err != nil {
		return nil
	} else {
		return req
	}
}

func DoAndLog(req *http.Request) (*http.Response, error) {
	return _do(req, true)
}

func Do(req *http.Request) (*http.Response, error) {
	return _do(req, false)
}

func _do(req *http.Request, log bool) (*http.Response, error) {
	logStruct := LogStruct{}
	bb, ok := req.Body.(buf.ByteBuf)
	if !ok {
		bb = buf.EmptyByteBuf()
		if req.Body != nil {
			bb.WriteReader(req.Body)
			req.Body.Close()
		}
	}

	bbl := len(bb.Bytes())
	if bbl < maxLogBodyBytes {
		logStruct.Request.Body = string(bb.Bytes())
	} else {
		logStruct.Request.Body = string(bb.Bytes()[:maxLogRequestBodyBytes])
	}

	if req.Body != nil {
		req.Body = bb
	}

	response, err := _client.Do(req)

	if log {
		logStruct.Request.BodyLength = bbl
		logStruct.Error = err
		logStruct.Uri = req.URL.RequestURI()
		logStruct.Request.URI = req.URL.RequestURI()
		logStruct.Request.Method = req.Method
		logStruct.Request.HOST = req.Host
		logStruct.Request.Headers = map[string]any{}
	}

	for name, value := range req.Header {
		valStr := ""
		if len(value) > 1 {
			for i := range value {
				if i == 0 {
					valStr = value[0]
				} else {
					valStr = fmt.Sprintf("%s;%s", valStr, value[i])
				}
			}
		} else {
			valStr = value[0]
		}

		if log {
			if name == "Authorization" {
				sha := sha256.New()
				sha.Write([]byte(valStr))
				logStruct.Request.Headers[name] = base64.RawURLEncoding.EncodeToString(sha.Sum(nil))
			} else {
				logStruct.Request.Headers[name] = valStr
			}
		}
	}

	if response != nil {
		if log {
			logStruct.Response.StatusCode = response.StatusCode
			logStruct.Response.Headers = map[string]any{}
			for name, value := range response.Header {
				valStr := ""
				if len(value) > 1 {
					for i := range value {
						if i == 0 {
							valStr = value[0]
						} else {
							valStr = fmt.Sprintf("%s;%s", valStr, value[i])
						}
					}
				} else {
					valStr = value[0]
				}

				logStruct.Response.Headers[name] = valStr
			}
		}

		bb := buf.EmptyByteBuf().WriteReader(response.Body)
		_ = response.Body.Close()
		response.Body = bb
		if log {
			logStruct.Response.OutBodyLength = len(bb.Bytes())
			if logStruct.Response.OutBodyLength < maxLogBodyBytes {
				logStruct.Response.Body = string(bb.Bytes())
			} else {
				logStruct.Response.Body = string(bb.Bytes()[:maxLogBodyBytes])
			}
		}
	}

	if logStruct.Request.BodyLength+logStruct.Response.OutBodyLength > maxLogBodyBytes {
		if logStruct.Request.BodyLength > maxLogBodyTrimBytes {
			logStruct.Request.Body = logStruct.Request.Body[:maxLogBodyTrimBytes]
		}

		if logStruct.Response.OutBodyLength > maxLogBodyTrimBytes {
			logStruct.Response.Body = logStruct.Response.Body[:maxLogBodyTrimBytes]
		}
	}

	if log {
		kklogger.DebugJ("httpclient:HttpClient.DoAndLog", logStruct)
	}

	return response, err
}

// LogStruct captures request and response details for structured logging.
type LogStruct struct {
	Uri      string                  `json:"uri"`
	Request  ghttp.RequestLogStruct  `json:"request"`
	Response ghttp.ResponseLogStruct `json:"response"`
	Error    error                   `json:"error,omitempty"`
}
