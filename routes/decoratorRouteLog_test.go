//   Copyright 2017 MSolution.IO
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"testing/quick"

	"github.com/asifbg-systems/jsonlog"
)

type basicLogMessage struct {
	Level   string           `json:"level"`
	Message string           `json:"message"`
	Data    basicRequestData `json:"data"`
}

type basicRequestData struct {
	Protocol  string    `json:"protocol"`
	Address   string    `json:"address"`
	UserAgent [1]string `json:"userAgent"`
}

func basicRequestDataExpected(brd basicRequestData) basicLogMessage {
	return basicLogMessage{
		Level:   "info",
		Message: "Received request.",
		Data:    brd,
	}
}

func TestRouteLogging(t *testing.T) {
	h := H(getFoo).With(
		RouteLog{},
	)
	buf := bytes.NewBuffer(make([]byte, 0x1000))
	f := func(brd basicRequestData) basicLogMessage {
		buf.Reset()
		logger := jsonlog.DefaultLogger.WithWriter(buf)
		request := httptest.NewRequest(http.MethodGet, "/", nil)
		request.Proto = brd.Protocol
		request.Header["User-Agent"] = brd.UserAgent[:]
		request.RemoteAddr = brd.Address
		request = request.WithContext(jsonlog.ContextWithLogger(request.Context(), logger))
		response := httptest.NewRecorder()
		s, r := h.Func(response, request, Arguments{})
		if s != http.StatusOK {
			t.Errorf("Status code should be %d, is %d instead.", http.StatusOK, s)
		}
		if rt, ok := r.(string); ok {
			if rt != getFooResponse {
				t.Errorf(
					"Response should be '%s', is '%s' instead.",
					getFooResponse,
					rt,
				)
			}
		} else {
			t.Errorf("Response should be %[1]T %#[1]v, is %[2]T %#[2]v instead.", getFooResponse, r)
		}
		var blm basicLogMessage
		if err := json.Unmarshal(buf.Bytes(), &blm); err != nil {
			t.Errorf("Failed to unmarshal log message with '%s'.", err.Error())
		}
		return blm
	}
	if err := quick.CheckEqual(basicRequestDataExpected, f, nil); err != nil {
		t.Error(err)
	}
}
