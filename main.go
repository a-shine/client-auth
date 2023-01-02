package main

import (
	"log"
	"net/http"
)

func main() {
	initDb()
	initCache()

	// we will implement these handlers in the next sections
	http.HandleFunc("/signin", Signin)
	http.HandleFunc("/signup", Signup)
	http.HandleFunc("/refresh", Refresh)
	http.HandleFunc("/logout", Logout)

	// start the server on port 8000
	log.Fatal(http.ListenAndServe(":8000", nil))
}
