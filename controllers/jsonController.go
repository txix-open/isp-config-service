package controllers

/*var HandlerSwagger = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Method not received"))
	}
	method := values.Get("method")
	if method == "" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Method not received"))
	}
	proxyRequestHandle(w, r, method)
})
*/
