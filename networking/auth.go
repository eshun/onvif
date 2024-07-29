package networking

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var (
	connectTimeOut   = time.Duration(10 * time.Second)
	readWriteTimeout = time.Duration(20 * time.Second)
	userAgent        = "AtScale"
)

const (
	nc = "00000001"
)

func NewRestResponse(body []byte, code int, err error) RestResponse {
	return RestResponse{
		RespBody:   body,
		StatusCode: code,
		Err:        err,
	}
}

func NewRestResponseEx(body []byte, code int, err error, req string) RestResponse {
	return RestResponse{
		RespBody:   body,
		StatusCode: code,
		Err:        err,
		ReqString:  req,
	}
}

func makeReqString(method, url, body string) string {
	return fmt.Sprintf("<%v> [%v] -- %v", method, url, body)
}

type RestResponse struct {
	ReqString  string
	RespBody   []byte
	StatusCode int
	Err        error
}

type RestClientAuth struct {
	user string
	pwd  string
}

func (this *RestClientAuth) SetAuth(user, pwd string) {
	this.user = user
	this.pwd = pwd
}

func (this *RestClientAuth) Request(method string, url string, reqBody []byte, headers map[string]string, timeout time.Duration, encryptionType string) RestResponse {
	index := strings.Index(url[9:], "/")
	uri := url[index+9:]
	return Auth(this.user, this.pwd, url, uri, method, reqBody, &headers, timeout, encryptionType)
}

func Request(method, url string, reqBody []byte, pHeaders map[string]string, cli *http.Client) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}

	hosts:=strings.Split(req.Host, ":")
	headers := http.Header{
		//"User-Agent":      []string{userAgent},
		//"Accept":          []string{"*/*"},
		//"Accept-Encoding": []string{"identity"},
		"Connection":      []string{"Keep-Alive"},
		"Host":            []string{hosts[0]},
	}
	if pHeaders != nil {
		for k,v := range pHeaders {
			headers[k] = []string{v}
		}
	}
	req.Header = headers

	resp, err := cli.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func Auth(username string, password string, url, uri string, method string, requestBody []byte, headerMap *map[string]string, timeout time.Duration, encryptionType string) RestResponse {
	client := NewTimeoutClient(timeout, timeout)
	jar := &myjar{}
	jar.jar = make(map[string][]*http.Cookie)
	client.Jar = jar

	resp, err := Request(method, url, requestBody, *headerMap, client)
	if err != nil {
		return NewRestResponse(nil, -1, err)
	}
	var respBody []byte = nil
	respBody, _ = ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized {
		(*headerMap)["Authorization"] = digistResponse(resp, username, password, method, uri, encryptionType)
		resp,err = Request(method, url, requestBody, *headerMap, client)
		if err != nil {
			return NewRestResponse(nil, -1, err)
		}
		respBody, _ = ioutil.ReadAll(resp.Body)
		defer resp.Body.Close()
	}

	return NewRestResponse(respBody, resp.StatusCode, err)
}

func digistResponse(resp *http.Response, username, password, method, uri string, encryptionType string) string {
	var authorization map[string]string = DigestAuthParams(resp)
	realmHeader := authorization["realm"]
	qopHeader := authorization["qop"]
	nonceHeader := authorization["nonce"]
	opaqueHeader := authorization["opaque"]
	algorithm := authorization["algorithm"]
	if algorithm == "" {
		algorithm = encryptionType
	}
	realm := realmHeader
	//encryptionType = algorithm

	// A1
	A1 := fmt.Sprintf("%s:%s:%s", username, realm, password)
	HA1 := H(A1, encryptionType)
	fmt.Println(fmt.Sprintf("A1: %v ---- %v", A1, HA1))

	// A2
	A2 := fmt.Sprintf("%v:%s", method, uri)
	HA2 := H(A2, encryptionType)
	fmt.Println(fmt.Sprintf("A2: %v ---- %v", A2, HA2))

	// response
	//cnonce := RandomKey()
	cnonce := "BC0688F75EF7B24CDE63418603E18D9E"
	response := H(strings.Join([]string{HA1, nonceHeader, nc, cnonce, qopHeader, HA2}, ":"), encryptionType)
	fmt.Println(fmt.Sprintf("response: %v ---- %v", strings.Join([]string{HA1, nonceHeader, nc, cnonce, qopHeader, HA2}, ":"), response))

	// now make header
	AuthHeader := fmt.Sprintf(`Digest username="%s", realm="%s", nonce="%s", uri="%s", algorithm=%v, response="%s", opaque="%s", qop=%s, nc=%s, cnonce="%s"`,
			username, realmHeader, nonceHeader, uri, algorithm, response, opaqueHeader, qopHeader, nc, cnonce)
	return AuthHeader
}

/*
 Parse Authorization header from the http.Request. Returns a map of
 auth parameters or nil if the header is not a valid parsable Digest
 auth header.
*/
func DigestAuthParams(r *http.Response) map[string]string {
	s := strings.SplitN(r.Header.Get("WWW-Authenticate"), " ", 2)
	if len(s) != 2 || s[0] != "Digest" {
		return nil
	}

	result := map[string]string{}
	for _, kv := range strings.Split(s[1], ",") {
		parts := strings.SplitN(kv, "=", 2)
		if len(parts) != 2 {
			continue
		}
		result[strings.Trim(parts[0], "\" ")] = strings.Trim(parts[1], "\" ")
	}
	return result
}

func RandomKey() string {
	k := make([]byte, 8)
	for bytes := 0; bytes < len(k); {
		n, err := rand.Read(k[bytes:])
		if err != nil {
			panic("rand.Read() failed")
		}
		bytes += n
	}
	return base64.StdEncoding.EncodeToString(k)
}

func H(data string, encryptionType string) string {
	switch encryptionType {
	case "MD5":
		return MD5(data)
	case "SHA256","SHA-256":
		return SHA256(data)
	}

	return ""
}

/*
 H function for MD5 algorithm (returns a lower-case hex MD5 digest)
*/
func MD5(data string) string {
	digest := md5.New()
	digest.Write([]byte(data))
	return hex.EncodeToString(digest.Sum(nil))
}

func SHA256(data string) string {
	digest := sha256.New()
	digest.Write([]byte(data))
	return hex.EncodeToString(digest.Sum(nil))
}

func timeoutDialer(cTimeout time.Duration, rwTimeout time.Duration) func(net, addr string) (c net.Conn, err error) {
	return func(netw, addr string) (net.Conn, error) {
		conn, err := net.DialTimeout(netw, addr, cTimeout)
		if err != nil {
			return nil, err
		}
		if rwTimeout > 0 {
			conn.SetDeadline(time.Now().Add(rwTimeout))
		}
		return conn, nil
	}
}

// apps will set three OS variables:
func NewTimeoutClient(cTimeout time.Duration, rwTimeout time.Duration) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
			DisableKeepAlives: false,
			Dial:              timeoutDialer(cTimeout, rwTimeout),
		},
	}
}

func DefaultTimeoutClient() *http.Client {
	return NewTimeoutClient(connectTimeOut, readWriteTimeout)
}

type myjar struct {
	jar map[string][]*http.Cookie
}

func (p *myjar) SetCookies(u *url.URL, cookies []*http.Cookie) {
	p.jar[u.Host] = cookies
}

func (p *myjar) Cookies(u *url.URL) []*http.Cookie {
	return p.jar[u.Host]
}
