package server

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
)

// Server responds to HTTP requests
type Server interface {
	// Close shuts the server down. If Close has already
	// been called, or Open was never called, then Close
	// is a noop. This method returns an error type
	// only to conform to the io.Closer interface, the error
	// will always be nil
	Close() error

	// GetGETRequests retrieves requests for
	// the given key where key is "path?query"
	GetGETRequests(key string) []http.Request

	// GetPOSTRequests retrieves requests for
	// for the given key where key is "path?query body"
	// body is expected to be an Multipart Post body with
	// a file named "file"
	GetPOSTRequests(key string) []http.Request

	// Open starts the server
	Open() error

	// Reset clears all requests and responses. This
	// should be called between every test to prevent
	// tests from affecting each other.
	Reset()

	// SetGETResponse sets the string response
	// for the given key where key is "path?query"
	// The response will automatically be an HTTP 200
	// and Content-Type application/json
	SetGETResponseBody(key, body string)

	// SetPOSTResponseBody sets the string response
	// for the given key where key is "path?query body"
	// body is expected to be an Multipart Post body with
	// a file named "file". The response will automatically
	// be an HTTP 200 and Content-Type application/json
	SetPOSTResponseBody(key, body string)

	// URL returns the url where the server can be found
	URL() url.URL
}

type _Server struct {
	server *httptest.Server
	url    *url.URL

	httpGETRequests   map[string][]http.Request
	httpGETResponses  map[string]_Response
	httpPOSTRequests  map[string][]http.Request
	httpPOSTResponses map[string]_Response
}

type _Response struct {
	StatusCode int
	Body       string
}

// New constructs an instance of Server that uses
// httptest
func New() Server {
	return &_Server{}
}

func (s *_Server) Close() error {
	if s.server == nil {
		return nil
	}
	s.server.Close()
	return nil
}

func (s *_Server) GetGETRequests(key string) []http.Request {
	return s.httpGETRequests[key]
}

func (s *_Server) GetPOSTRequests(key string) []http.Request {
	return s.httpPOSTRequests[key]
}

func (s *_Server) Open() error {
	var err error

	s.server = httptest.NewServer(http.HandlerFunc(s.handleRequest))
	s.url, err = url.Parse(s.server.URL)
	return err
}

func (s *_Server) Reset() {
	s.httpGETResponses = map[string]_Response{}
	s.httpPOSTResponses = map[string]_Response{}

	s.httpGETRequests = map[string][]http.Request{}
	s.httpPOSTRequests = map[string][]http.Request{}
}

func (s *_Server) SetGETResponseBody(key, responseBody string) {
	s.httpGETResponses[key] = _Response{
		StatusCode: http.StatusOK,
		Body:       responseBody,
	}
}

func (s *_Server) SetPOSTResponseBody(key, responseBody string) {
	s.httpPOSTResponses[key] = _Response{
		StatusCode: http.StatusOK,
		Body:       responseBody,
	}
}

func (s *_Server) URL() url.URL {
	return *s.url
}

// privates
func (s *_Server) handleRequest(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleGetRequest(w, r)
		return
	case http.MethodPost:
		s.handlePostRequest(w, r)
		return
	}
}

func (s *_Server) handleGetRequest(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path + "?" + r.URL.RawQuery
	s.httpGETRequests[key] = append(s.httpGETRequests[key], *r)

	response, ok := s.httpGETResponses[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("No httpGETResponse for '%v'", key)))
		return
	}

	w.WriteHeader(response.StatusCode)
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte(response.Body))
}

func (s *_Server) handlePostRequest(w http.ResponseWriter, r *http.Request) {
	f, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	body, err := ioutil.ReadAll(f)
	if err != nil {
		http.Error(w, err.Error(), 500)
	}

	key := r.URL.Path + "?" + r.URL.RawQuery + " " + string(body)
	s.httpPOSTRequests[key] = append(s.httpPOSTRequests[key], *r)

	response, ok := s.httpPOSTResponses[key]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(fmt.Sprintf("No httpPOSTResponse for '%v'", key)))
		return
	}

	w.WriteHeader(response.StatusCode)
	w.Header().Add("Content-Type", "application/json")
	w.Write([]byte(response.Body))
}
