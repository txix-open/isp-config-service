package mux

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHTTP1(t *testing.T) {
	a := assert.New(t)
	path := "/test/"
	method := "GET"
	mockResData := []byte("948 res")
	var responseData []byte

	rootListener := newLocalListener()
	mux := New(rootListener)
	httpListener := mux.Match(HTTP1())
	mux.OnError(func(err error) {
		a.Fail("connection not matched by a http matcher")
	})

	go func() {
		err := mux.Serve()
		log.Println(err)
	}()

	handler := func(writer http.ResponseWriter, request *http.Request) {
		var err error
		a.Equal(method, request.Method)
		n, err := writer.Write(mockResData)
		a.NoError(err)
		a.Len(mockResData, n)
	}
	httpMux := http.NewServeMux()
	httpMux.HandleFunc(path, handler)
	httpServer := httptest.NewUnstartedServer(httpMux)
	httpServer.Listener = httpListener
	httpServer.Start()

	url := httpServer.URL + path
	httpClient := httpServer.Client()
	req, _ := http.NewRequest(method, url, nil)
	res, err := httpClient.Do(req)
	a.NoError(err)
	if a.NotNil(res) {
		a.Equal(200, res.StatusCode)
		responseData, err = ioutil.ReadAll(res.Body)
	}

	time.Sleep(10 * time.Millisecond)
	a.Equal(mockResData, responseData)
}

func TestAny(t *testing.T) {
	a := assert.New(t)
	mockReqData := []byte("1 somre")
	requestData := make([]byte, len(mockReqData))

	rootListener := newLocalListener()
	mux := New(rootListener)
	httpListener := mux.Match(HTTP1())
	tcpListener := mux.Match(Any())
	go func() {
		err := mux.Serve()
		log.Println(err)
	}()

	handler := func(writer http.ResponseWriter, request *http.Request) {
		a.Fail("matched http handler")
	}
	httpMux := http.NewServeMux()
	httpMux.HandleFunc("/", handler)
	httpServer := httptest.NewUnstartedServer(httpMux)
	httpServer.Listener = httpListener
	httpServer.Start()

	hadTcpConn := false
	go func() {
		for {
			conn, err := tcpListener.Accept()
			if hadTcpConn {
				a.Fail("more than one tcp conn")
				return
			}
			hadTcpConn = true
			a.NoError(err)
			_, err = conn.Read(requestData)
			a.NoError(err)
			a.Equal(mockReqData, requestData)
		}
	}()

	tcpAddr := tcpListener.Addr().String()
	conn, err := net.Dial("tcp", tcpAddr)
	a.NoError(err)
	n, err := conn.Write(mockReqData)
	a.Len(mockReqData, n)
	a.NoError(err)

	time.Sleep(10 * time.Millisecond)
	a.Equal(mockReqData, requestData)
	a.True(hadTcpConn)
}

func newLocalListener() net.Listener {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(fmt.Sprintf("failed to listen on a port: %v", err))
	}
	return l
}
