package main

// https://github.com/tryvium-travels/memongo
//https://medium.com/@victor.neuret/mocking-the-official-mongo-golang-driver-5aad5b226a78
// Good repo with examples on how to mock mongoDB

// Here we are testing the API endpoints as a form of integration testing. This
// requires other dependent services to be available such as the database and
// cache so docker compose is used to faciliate orchestration

// func TestRegisterNewUser(t *testing.T) {
// 	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
// 	defer mt.Close()

// 	mt.Run("success", func(mt *mtest.T) {
// 		users := mt.Coll
// 		rdb := getCache()

// 		// Get handler object
// 		handler := createHandler(users, rdb)

// 		// Create http recorder to record response
// 		recorder := httptest.NewRecorder()

// 		// Generate a new unique user email
// 		gofakeit.Seed(0)
// 		newUserEmail := gofakeit.Email()

// 		// Create a new user json
// 		user := `{"email": "` + newUserEmail + `", "password": "somePassword",
// 					"firstName": "John", "lastName": "Smith",
// 					"groups": ["admin"]}`

// 		// Create a new request
// 		req, _ := http.NewRequest("POST", "/register-user", strings.NewReader(user))

// 		// Send request to service
// 		handler.ServeHTTP(recorder, req)

// 		// Check if user is indeed inserted into the database
// 		filter := bson.D{{Key: "email", Value: newUserEmail}}
// 		mongoErr := users.FindOne(context.Background(), filter).Err()

// 		if mongoErr == mongo.ErrNoDocuments {
// 			t.Error("User was not added to database")
// 		}

// 		// Check if user creation response is sent to the client
// 		if recorder.Code != http.StatusCreated {
// 			t.Error("Endpoint did not return correct status code")
// 		}

// 		expectedUser := Client{
// 			Id:    primitive.NewObjectID(),
// 			Email: "john.doe@test.com",
// 		}

// 		mt.AddMockResponses(mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
// 			{"_id", expectedUser.Id},
// 			{"name", expectedUser.Name},
// 			{"email", expectedUser.Email},
// 		}))
// 		userResponse, err := getFromID(expectedUser.Id)
// 		assert.Nil(t, err)
// 		assert.Equal(t, &expectedUser, userResponse)
// 	})
// }

// Test case for successful user registration
// func TestRegisterNewUser(t *testing.T) {
// 	// Get MongoDB collection and Redis client
// 	users := usersCollection
// 	rdb := getCache()

// 	// Get handler object
// 	handler := createHandler(users, rdb)

// 	// Create http recorder to record response
// 	recorder := httptest.NewRecorder()

// 	// Generate a new unique user email
// 	gofakeit.Seed(0)
// 	newUserEmail := gofakeit.Email()

// 	// Create a new user json
// 	user := `{"email": "` + newUserEmail + `", "password": "somePassword", "firstName": "John", "lastName": "Smith", "groups": ["admin"]}`

// 	// Create a new request
// 	req, _ := http.NewRequest("POST", "/register-user", strings.NewReader(user))

// 	// Send request to service
// 	handler.ServeHTTP(recorder, req)

// 	// Check if user is indeed inserted into the database
// 	filter := bson.D{{Key: "email", Value: newUserEmail}}
// 	mongoErr := users.FindOne(context.Background(), filter).Err()

// 	if mongoErr == mongo.ErrNoDocuments {
// 		t.Error("User was not added to database")
// 	}

// 	// Check if user creation response is sent to the client
// 	if recorder.Code != http.StatusCreated {
// 		t.Error("Endpoint did not return correct status code")
// 	}
// }

// func TestSuccessfulServiceRegistration(t *testing.T) {
// 	// Get MongoDB collection and Redis client
// 	users := getUserCollection()
// 	rdb := getCache()

// 	// Get handler object
// 	handler := createHandler(users, rdb)

// 	// Create http recorder to record response
// 	recorder := httptest.NewRecorder()

// 	// Generate a new unique user email
// 	gofakeit.Seed(0)
// 	newUserEmail := gofakeit.Email()

// 	// Create a new user json
// 	servicePayload := `{"email": "` + newUserEmail + `", "name": "Service A", "groups": ["tempProbes"]}`

// 	// Create a new request
// 	req, _ := http.NewRequest("POST", "/register-service", strings.NewReader(servicePayload))

// 	// Send request to service
// 	handler.ServeHTTP(recorder, req)

// 	// Check if user is indeed inserted into the database
// 	filter := bson.D{{Key: "email", Value: newUserEmail}}
// 	mongoErr := users.FindOne(context.Background(), filter).Err()

// 	if mongoErr == mongo.ErrNoDocuments {
// 		t.Error("Service was not added to database")
// 	}

// 	// Check if user creation response is sent to the client
// 	if recorder.Code != http.StatusCreated {
// 		t.Error("Endpoint did not return correct status code")
// 	}

// 	fmt.Println(recorder.Body.String())
// }

// Test case for attempting to register a user that already exists
// func TestRegisterPreexistingUser(t *testing.T) {
// 	// Get MongoDB collection and Redis client
// 	users := getUserCollection()
// 	rdb := getCache()

// 	// Get handler object
// 	handler := createHandler(users, rdb)

// 	// Create http recorder to record response
// 	recorder := httptest.NewRecorder()

// 	// Generate a new unique user email
// 	gofakeit.Seed(0)
// 	newUserEmail := gofakeit.Email()

// 	// Pre-insert a user into the database (so that we can attempt to register it again)
// 	users.InsertOne(context.Background(), bson.D{
// 		{Key: "email", Value: newUserEmail},
// 		{Key: "password", Value: "somePasswordHash"},
// 		{Key: "firstName", Value: "John"},
// 		{Key: "lastName", Value: "Smith"},
// 	})

// 	// New user json with same email as the pre-inserted user
// 	user := `{"email": "` + newUserEmail + `", "password": "somePassword", "firstName": "John", "lastName": "Smith", "groups": ["admin"]}`

// 	// Create a new request
// 	req, _ := http.NewRequest("POST", "/register-user", strings.NewReader(user))

// 	// Send request to service
// 	handler.ServeHTTP(recorder, req)

// 	if recorder.Code != http.StatusConflict {
// 		t.Error("Endpoint did not return correct status code")
// 	}
// }

// Test case for attempting to register a user with an invalid payload
// func TestInvalidRegisterPayload(t *testing.T) {
// 	// Get MongoDB collection and Redis client
// 	users := getUserCollection()
// 	rdb := getCache()

// 	// Get handler object
// 	handler := createHandler(users, rdb)

// 	// Create http recorder to record response
// 	recorder := httptest.NewRecorder()

// 	// New user json with missing firstName field
// 	user := `{"email": "john@smith.com", "password": "somePassword", "lastName": "Smith", "groups": ["admin"]}`

// 	// New user json with misspelled email field
// 	// user = `{"emal": "john@smith.com", "password": "somePassword", "lastName": "Smith"}`

// 	// Create a new request
// 	req, _ := http.NewRequest("POST", "/register-user", strings.NewReader(user))

// 	// Send request to service
// 	handler.ServeHTTP(recorder, req)

// 	if recorder.Code != http.StatusBadRequest {
// 		t.Error("Endpoint did not return correct status code")
// 	}
// }
