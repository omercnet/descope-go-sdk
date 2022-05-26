package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"strings"
	"time"

	"github.com/descope/go-sdk/descope/errors"
	"github.com/descope/go-sdk/descope/logger"
	"github.com/descope/go-sdk/descope/utils"
)

const (
	defaultURL = "https://descope.com"
)

var (
	Routes = endpoints{
		version: "/v1/",
		auth: struct {
			signInOTP  string
			signUpOTP  string
			verifyCode string
		}{
			signInOTP:  "auth/signin/otp",
			signUpOTP:  "auth/signup/otp",
			verifyCode: "auth/code/verify",
		},
		logoutAll: "/logoutall",
		keys:      "/keys/",
		refresh:   "/refresh",
	}
)

type endpoints struct {
	version string
	auth    struct {
		signInOTP  string
		signUpOTP  string
		verifyCode string
	}
	logoutAll string
	keys      string
	refresh   string
}

func (e *endpoints) SignInOTP() string {
	return path.Join(e.version, e.auth.signInOTP)
}
func (e *endpoints) SignUpOTP() string {
	return path.Join(e.version, e.auth.signUpOTP)
}
func (e *endpoints) VerifyCode() string {
	return path.Join(e.version, e.auth.verifyCode)
}
func (e *endpoints) Logout() string {
	return path.Join(e.version, e.logoutAll)
}
func (e *endpoints) GetKeys() string {
	return path.Join(e.version, e.keys)
}
func (e *endpoints) RefreshToken() string {
	return path.Join(e.version, e.refresh)
}

type ClientParams struct {
	BaseURL              string
	DefaultClient        IHttpClient
	CustomDefaultHeaders map[string]string

	ProjectID string
}

type IHttpClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Client struct {
	httpClient IHttpClient
	uri        string
	headers    map[string]string
	conf       ClientParams
}
type HTTPResponse struct {
	Req     *http.Request
	Res     *http.Response
	BodyStr string
}
type HTTPRequest struct {
	Headers    map[string]string
	BaseURL    string
	ResBodyObj interface{}
	Cookies    []*http.Cookie
}

func NewClient(conf ClientParams) *Client {
	httpClient := conf.DefaultClient
	if httpClient == nil {
		t := http.DefaultTransport.(*http.Transport).Clone()
		t.MaxIdleConns = 100
		t.MaxConnsPerHost = 100
		t.MaxIdleConnsPerHost = 100

		httpClient = &http.Client{
			Timeout:   time.Second * 10,
			Transport: t,
		}
	}
	defaultHeaders := map[string]string{}

	for key, value := range conf.CustomDefaultHeaders {
		defaultHeaders[key] = value
	}

	if conf.BaseURL == "" {
		conf.BaseURL = defaultURL
	}

	return &Client{
		uri:        conf.BaseURL,
		httpClient: httpClient,
		headers:    defaultHeaders,
		conf:       conf,
	}
}

func (c *Client) DoGetRequest(uri string, options *HTTPRequest) (*HTTPResponse, error) {
	return c.DoRequest(http.MethodGet, uri, nil, options)
}

func (c *Client) DoPostRequest(uri string, body interface{}, options *HTTPRequest) (*HTTPResponse, error) {
	if options == nil {
		options = &HTTPRequest{}
	}
	if options.Headers == nil {
		options.Headers = map[string]string{}
	}
	if _, ok := options.Headers["Content-Type"]; !ok {
		options.Headers["Content-Type"] = "application/json"
	}

	var payload io.Reader
	if body != nil {
		if b, err := utils.Marshal(body); err == nil {
			payload = bytes.NewBuffer(b)
		} else {
			return nil, err
		}
	}

	return c.DoRequest(http.MethodPost, uri, payload, options)
}

func (c *Client) DoRequest(method, uriPath string, body io.Reader, options *HTTPRequest) (*HTTPResponse, error) {
	if options == nil {
		options = &HTTPRequest{}
	}

	base := c.uri
	if options.BaseURL != "" {
		base = options.BaseURL
	}

	url := fmt.Sprintf("%s/%s", base, strings.TrimLeft(uriPath, "/"))
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	for key, value := range c.headers {
		req.Header.Add(key, value)
	}

	for key, value := range options.Headers {
		req.Header.Add(key, value)
	}
	for _, cookie := range options.Cookies {
		req.AddCookie(cookie)
	}

	req.SetBasicAuth(c.conf.ProjectID, "")

	logger.LogDebug("sending request to [%s]", url)
	response, err := c.httpClient.Do(req)
	if err != nil {
		logger.LogInfo("failed sending request to [%s]", url)
		return nil, err
	}

	if response.Body != nil {
		defer response.Body.Close()
	}
	if !isResponseOK(response) {
		err = c.parseResponseError(response)
		logger.LogDebug("failed sending request to [%s] with [%s]", url, err)
		return nil, err
	}

	resBytes, err := c.parseBody(response)
	if err != nil {
		return nil, err
	}

	if options.ResBodyObj != nil {
		if err = utils.Unmarshal(resBytes, &options.ResBodyObj); err != nil {
			return nil, err
		}
	}

	return &HTTPResponse{
		Req:     req,
		Res:     response,
		BodyStr: string(resBytes),
	}, nil
}

func (c *Client) parseBody(response *http.Response) (resBytes []byte, err error) {
	if response.Body != nil {
		resBytes, err = ioutil.ReadAll(response.Body)
		if err != nil {
			logger.LogInfo("failed reading body from request to [%s]", response.Request.URL.String())
			return nil, err
		}
	}
	return
}

func (c *Client) parseResponseError(response *http.Response) error {
	if response.StatusCode == http.StatusUnauthorized {
		return errors.NewUnauthorizedError()
	}
	if response.StatusCode == http.StatusNotFound {
		return errors.NewError("404", fmt.Sprintf("url [%s] not found", response.Request.URL.String()))
	}

	body, err := c.parseBody(response)
	if err != nil {
		return err
	}

	var responseErr *errors.WebError
	if err := json.Unmarshal(body, &responseErr); err != nil {
		logger.LogInfo("failed to load error from response [error: %s]", err)
		return errors.NewValidationError(string(body))
	}
	return responseErr
}

func isResponseOK(response *http.Response) bool {
	return response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices
}