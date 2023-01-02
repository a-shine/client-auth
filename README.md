# Simple user authentication/authorization service

Simple user authentication service providing functionality to register, login (generate JWT auth tokens)... users.

The service stores registered users in a NoSQL MongoDB database (no need for relational as there are no relationships to capture).

The authentication service provides an `isAuth/` route and can hence be used with the a-shine/api-gateway. 

User authentication is verified with JWT tokens, usually this form of authentication is statless i.e. doesn't require querying a database to verify if a user is authenticated but this is insecure when having ti deactivate users or carry out requests by deleted users. An option is to cache a blacklist of invalid tokens (maybe using Redis) but here we simply query the user in the MongoDB database to verify if he is authenticated and authorised. This can be done in a complexity of O(1) as we query users by ID.

Dependencies MongoDB


curl -v -XPOST -H "Content-type: application/json" -d '{"password": "test", "email":"sjchgsajdhgc", "first_name":"alex", "last_name":"kjsd"}' 'localhost:8000/signup'

