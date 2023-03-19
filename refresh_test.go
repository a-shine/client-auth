package main

// func TestRefresh(t *testing.T) {
// 	// Get MongoDB collection and Redis client
// 	users := getUserCollection()
// 	rdb := getCache()

// 	// Get handler object
// 	handler := createHandler(users, rdb)

// 	// Create http recorder to record response
// 	recorder := httptest.NewRecorder()

// 	email := randomEmail()
// 	hashedPass, _ := hashAndSalt("somePassword")

// 	// Insert a new user into the database
// 	users.InsertOne(context.Background(), bson.D{
// 		{Key: "email", Value: email},
// 		{Key: "hashedPassword", Value: hashedPass},
// 		{Key: "firstName", Value: "John"},
// 		{Key: "lastName", Value: "Smith"},
// 	})

// 	// Create a new user json
// 	login := `{"email": "` + email + `", "password": "somePassword"}`

// 	// Create a new request
// 	req1, _ := http.NewRequest("POST", "/login", strings.NewReader(login))

// 	// Send request to service
// 	handler.ServeHTTP(recorder, req1)

// 	tokenCookies := recorder.Result().Cookies()

// 	fmt.Println(tokenCookies)

// 	// make a new request with the cookie
// 	req2, _ := http.NewRequest("POST", "/refresh", nil)

// 	for _, cookie := range tokenCookies {
// 		req2.AddCookie(cookie)
// 	}
// 	// req2.AddCookie(tokenCookies)

// 	fmt.Println(req2.Cookies())

// 	handler.ServeHTTP(recorder, req2)

// 	fmt.Println(recorder.Body)
// 	fmt.Println(recorder.Code)

// 	if recorder.Code != http.StatusOK {
// 		t.Error("Endpoint did not return correct status code")
// 	}
// }
