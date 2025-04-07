package function

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/spf13/cast"
)

const (
	botToken        = "6339787602:AAHTlrZ7wT7h_AU-KXzkpX3O06vg0NYlAmo"
	chatID          = "-162256495"
	baseUrl         = "https://api.admin.u-code.io"
	logFunctionName = "ucode-template"
	IsHTTP          = true // if this is true banchmark test works.
)

// Request structures
type (
	// Handle request body
	NewRequestBody struct {
		RequestData HttpRequest `json:"request_data"`
		Auth        AuthData    `json:"auth"`
		Data        Data        `json:"data"`
	}

	HttpRequest struct {
		Method  string      `json:"method"`
		Path    string      `json:"path"`
		Headers http.Header `json:"headers"`
		Params  url.Values  `json:"params"`
		Body    []byte      `json:"body"`
	}

	AuthData struct {
		Type string                 `json:"type"`
		Data map[string]interface{} `json:"data"`
	}

	// Function request body >>>>> GET_LIST, GET_LIST_SLIM, CREATE, UPDATE
	Request struct {
		Data map[string]interface{} `json:"data"`
	}

	// most common request structure -> UPDATE, MULTIPLE_UPDATE, CREATE, DELETE
	Data struct {
		AppId      string                 `json:"app_id"`
		Method     string                 `json:"method"`
		ObjectData map[string]interface{} `json:"object_data"`
		ObjectIds  []string               `json:"object_ids"`
		TableSlug  string                 `json:"table_slug"`
		UserId     string                 `json:"user_id"`
	}

	FunctionRequest struct {
		BaseUrl     string  `json:"base_url"`
		TableSlug   string  `json:"table_slug"`
		AppId       string  `json:"app_id"`
		Request     Request `json:"request"`
		DisableFaas bool    `json:"disable_faas"`
	}
)

// Response structures
type (
	// Create function response body >>>>> CREATE
	Datas struct {
		Data struct {
			Data struct {
				Data map[string]interface{} `json:"data"`
			} `json:"data"`
		} `json:"data"`
	}

	// ClientApiResponse This is get single api response >>>>> GET_SINGLE_BY_ID, GET_SLIM_BY_ID
	ClientApiResponse struct {
		Data ClientApiData `json:"data"`
	}

	ClientApiData struct {
		Data ClientApiResp `json:"data"`
	}

	ClientApiResp struct {
		Response map[string]interface{} `json:"response"`
	}

	Response struct {
		Status string                 `json:"status"`
		Data   map[string]interface{} `json:"data"`
	}

	// GetListClientApiResponse This is get list api response >>>>> GET_LIST, GET_LIST_SLIM
	GetListClientApiResponse struct {
		Data GetListClientApiData `json:"data"`
	}

	GetListClientApiData struct {
		Data GetListClientApiResp `json:"data"`
	}

	GetListClientApiResp struct {
		Response []map[string]interface{} `json:"response"`
	}

	// ClientApiUpdateResponse This is single update api response >>>>> UPDATE
	ClientApiUpdateResponse struct {
		Status      string `json:"status"`
		Description string `json:"description"`
		Data        struct {
			TableSlug string                 `json:"table_slug"`
			Data      map[string]interface{} `json:"data"`
		} `json:"data"`
	}

	// ClientApiMultipleUpdateResponse This is multiple update api response >>>>> MULTIPLE_UPDATE
	ClientApiMultipleUpdateResponse struct {
		Status      string `json:"status"`
		Description string `json:"description"`
		Data        struct {
			Data struct {
				Objects []map[string]interface{} `json:"objects"`
			} `json:"data"`
		} `json:"data"`
	}

	ResponseStatus struct {
		Status string `json:"status"`
	}
)

// Testing types
type (
	Asserts struct {
		Request  NewRequestBody
		Response Response
	}

	FunctionAssert struct{}
)

func (f FunctionAssert) GetAsserts() []Asserts {
	var appId = os.Getenv("APP_ID")

	return []Asserts{
		{
			Request: NewRequestBody{
				Data: Data{
					AppId:     appId,
					ObjectIds: []string{"96b6c9e0-ec0c-4297-8098-fa9341c40820"},
				},
			},
			Response: Response{
				Status: "done",
			},
		},
		{
			Request: NewRequestBody{
				Data: Data{
					AppId:     appId,
					ObjectIds: []string{"96b6c9e0-ec0c-4297-8098"},
				},
			},
			Response: Response{Status: "error"},
		},
	}
}

func (f FunctionAssert) GetBenchmarkRequest() Asserts {
	var appId = os.Getenv("APP_ID")
	return Asserts{
		Request: NewRequestBody{
			Data: Data{
				AppId:     appId,
				ObjectIds: []string{"96b6c9e0-ec0c-4297-8098-fa9341c40820"},
			},
		},
		Response: Response{
			Status: "done",
		},
	}
}

// Handle a serverless request
func Handle(req []byte) string {
	Send(string(req))
	var (
		response Response
		request  NewRequestBody
	)

	err := json.Unmarshal(req, &request)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling request", "error": err.Error()}
		response.Status = "error"
		responseByte, _ := json.Marshal(response)
		return string(responseByte)
	}
	Send(fmt.Sprintf("%v", request))

	if request.Data.ObjectData["user_id"] != nil {
		client, _, err := GetSlimObject(FunctionRequest{
			BaseUrl:   urlConst,
			TableSlug: "cleints",
			AppId:     appId,
			Request: Request{
				Data: map[string]interface{}{
					"guid": fmt.Sprintf("%v", request.Data.ObjectData["user_id"]),
				}}),
		)
		if err != nil {
			Send("IN madadio-generate-balance-id" + err.Error())
			response.Data = map[string]interface{}{"message": "Error while getting slim object", "error": err.Error()}
			response.Status = "error"
			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}
		Send(fmt.Sprintf("%v", client))

		if client.Data.Data.Response["balance_id"] != nil {
			response.Data = map[string]interface{}{ }
			response.Status = "done"
			responseByte, _ := json.Marshal(response)
			return string(responseByte)
		}
	}
}

func generateSevenDigitNumber() int {
	rand.Seed(time.Now().UnixNano())
	num := rand.Intn(9000000) + 1000000
	return num
}

func UpdateObject(in FunctionRequest) (ClientApiUpdateResponse, Response, error) {
	response := Response{
		Status: "done",
	}

	var updateObject ClientApiUpdateResponse
	updateObjectResponseInByte, err := DoRequest(fmt.Sprintf("%s/v1/object/%s?from-ofs=%t", in.BaseUrl, in.TableSlug, in.DisableFaas), "PUT", in.Request, in.AppId)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while updating object", "error": err.Error()}
		response.Status = "error"
		return ClientApiUpdateResponse{}, response, errors.New("error")
	}

	err = json.Unmarshal(updateObjectResponseInByte, &updateObject)
	if err != nil {
		response.Data = map[string]interface{}{"message": "Error while unmarshalling update object", "error": err.Error()}
		response.Status = "error"
		return ClientApiUpdateResponse{}, response, errors.New("error")
	}

	return updateObject, response, nil
}
