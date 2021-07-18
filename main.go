package main

import (
	"context"
	"time"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
  "net/http"
	"fmt"
	"log"
	"github.com/gorilla/mux"
	"io/ioutil"
	"encoding/json"
	"regexp"
)

type AppUserCreds struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

var connectionString = "mongodb://localhost"


func changePassword(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
		reqBody, err := ioutil.ReadAll(r.Body)
		var appUserCred AppUserCreds

		client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)
 
 	  appUsersDatabase := client.Database("appUsers")
    appUserLoginsCollection := appUsersDatabase.Collection("appUserLogins")

    err = json.Unmarshal(reqBody, &appUserCred)
    if err != nil {
        panic(err)
    }

		err = validatePassword(appUserCred.Password)
    if err != nil {
				w.Write([]byte(`{"message": "` + err.Error() + `"}`))
				return
    } 

		var existingUser bson.M
		if err = appUserLoginsCollection.FindOne(ctx, bson.M{"username": appUserCred.Username}).Decode(&existingUser); err != nil {
				fmt.Println("User doesn't exist. New user will be created.")				
		}
				
		if existingUser["password"] == appUserCred.Password {
				w.Write([]byte(`{"message": "New password cannot be the same as old password"}`))
				return
		}

		if existingUser != nil {
				result, err := appUserLoginsCollection.UpdateOne(
						ctx,
						bson.M{"username": appUserCred.Username},
						bson.D{
								{"$set", bson.D{{"password", appUserCred.Password}}},
						},
				)
				if err != nil {
						log.Fatal(err)
				}
				fmt.Println("Updated %v Documents!", result.ModifiedCount)
				w.Write([]byte(`{"message": "Password updated against existing user"}`))
		} else {
				appUserLoginsResult, err := appUserLoginsCollection.InsertOne(ctx, bson.D{
				{"username", appUserCred.Username},
				{"password", appUserCred.Password},
				{"updatedon", time.Now()},
				})
				if err != nil {
						log.Fatal(err)
				}
				fmt.Println("New user created with %+v", appUserLoginsResult.InsertedID)
				w.Write([]byte(`{"message": "New user created with given password"}`))
		}
}

func fetchAllUsers(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)

		client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)
 	  appUsersDatabase := client.Database("appUsers")
    appUserLoginsCollection := appUsersDatabase.Collection("appUserLogins")
		opts := options.Find()
		opts.SetSort(bson.D{{"updatedon", -1}})
		sortCursor, err := appUserLoginsCollection.Find(ctx, bson.D{}, opts)
		if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
    		return
		}
		var appUsersList []bson.M
		if err = sortCursor.All(ctx, &appUsersList); err != nil {
				log.Fatal(err)
		}
		json.NewEncoder(w).Encode(appUsersList)
}

func deleteAllUsers(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
		
		client, err := mongo.NewClient(options.Client().ApplyURI(connectionString))
    if err != nil {
        log.Fatal(err)
    }
    ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
    err = client.Connect(ctx)
    if err != nil {
        log.Fatal(err)
    }
    defer client.Disconnect(ctx)
 	  appUsersDatabase := client.Database("appUsers")
    appUserLoginsCollection := appUsersDatabase.Collection("appUserLogins")

		err1 := appUserLoginsCollection.Drop(ctx);
		if err1 != nil {
				log.Fatal(err1)
		}
		w.Write([]byte(`{"message": "Deleted all"}`))
}

func validatePassword(ps string) error {
		if len(ps) < 8 {
				return fmt.Errorf("Password should have at least 8 characters")
		}
		if len(ps) > 16 {
				return fmt.Errorf("Password shouldn't be longer than 16 characters")
		}
		num := `[0-9]{1}`
		a_z := `[a-z]{1}`
		A_Z := `[A-Z]{1}`
		symbol := `[!@#~$%^&*()+|_]{1}`
		if b, err := regexp.MatchString(num, ps); !b || err != nil {
				return fmt.Errorf("Password needs to have at least one number")
		}
		if b, err := regexp.MatchString(a_z, ps); !b || err != nil {
				return fmt.Errorf("Password needs to have at least one lower case letter")
		}
		if b, err := regexp.MatchString(A_Z, ps); !b || err != nil {
				return fmt.Errorf("Password needs to have at least one upper case letter")
		}
		if b, err := regexp.MatchString(symbol, ps); !b || err != nil {
				return fmt.Errorf("Password needs to have at least one special character")
		}
		return nil
}

func handleRequests() {
    r := mux.NewRouter()
    api := r.PathPrefix("").Subrouter()
    api.HandleFunc("/changepassword", changePassword).Methods(http.MethodPost)
    api.HandleFunc("/fetchallusers", fetchAllUsers).Methods(http.MethodGet)
		api.HandleFunc("/deleteallusers", deleteAllUsers).Methods(http.MethodGet)
    log.Fatal(http.ListenAndServe(":8080", r))
}

func main() {
		handleRequests()
}
