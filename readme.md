This system is used to reset user's password after some predefined validations. 

To start the application:
```
go mod init go-password-reset
go mod tidy
go run main.go
```

You may use the postman collections shared in the same repo to test the APIs. 


Pending:
1. Validation for username
2. Separation of DB connection from APIs
3. Addition of a config file to hold creds and other variables
4. Modularize the code
5. Allow to delete specific users
6. Add limit offset to Fetch Users API
7. Change delete API to POST or DELETE
8. Standardize HTTP responses
9. Enhance the readme doc

