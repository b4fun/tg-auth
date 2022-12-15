package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"time"
)

func dumpReqest(req *http.Request) {
	b, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Printf("dump output failed: %s\n", err)
		return
	}
	fmt.Println(string(b))
}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		dumpReqest(r)

		fmt.Println("incoming request - root")

		w.Write([]byte("hello, world"))
	})

	http.HandleFunc("/test/", func(w http.ResponseWriter, r *http.Request) {
		dumpReqest(r)

		if r.Header.Get("foo") != "" {
			http.Redirect(w, r, "https://google.com", 302)
			return
		}

		fmt.Println("incoming request")

		cookie := &http.Cookie{Name: "foo", Value: "bar", Expires: time.Now().Add(time.Hour)}
		fmt.Println("cookie to write", cookie.String())
		http.SetCookie(w, cookie)
		w.Write([]byte("hello, world"))
	})

	http.ListenAndServe(":8082", nil)
}
