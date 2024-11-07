package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
)

// Handle logging activity
func logActivity(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tokenString := r.Header.Get("Authorization")
	if len(tokenString) < 8 {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString[7:], claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	var activity Activity
	if err := json.NewDecoder(r.Body).Decode(&activity); err != nil {
		sendErrorResponse(w, http.StatusBadRequest, "Invalid input")
		return
	}

	activity.Username = claims.Username
	activity.Timestamp = time.Now()
	activitiesCollection := client.Database("your_database").Collection("activities")
	_, err = activitiesCollection.InsertOne(context.TODO(), activity)
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to log activity")
		return
	}

	log.Println("Activity logged:", activity)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
}

// Fetch activity feed
func getActivityFeed(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tokenString := r.Header.Get("Authorization")
	if len(tokenString) < 8 {
		sendErrorResponse(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	claims := &Claims{}
	_, err := jwt.ParseWithClaims(tokenString[7:], claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		sendErrorResponse(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	activitiesCollection := client.Database("your_database").Collection("activities")
	cursor, err := activitiesCollection.Find(context.TODO(), bson.M{"username": claims.Username})
	if err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to fetch activities")
		return
	}
	defer cursor.Close(context.TODO())

	var activities []Activity
	if err := cursor.All(context.TODO(), &activities); err != nil {
		sendErrorResponse(w, http.StatusInternalServerError, "Failed to decode activities")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(activities)
	log.Println("Fetched activities:", activities)
}

func sendErrorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// // Handle logging activity
// func logActivity(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	tokenString := r.Header.Get("Authorization")
// 	claims := &Claims{}

// 	_, err := jwt.ParseWithClaims(tokenString[7:], claims, func(token *jwt.Token) (interface{}, error) {
// 		return jwtSecret, nil
// 	})
// 	if err != nil {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	var activity Activity
// 	if err := json.NewDecoder(r.Body).Decode(&activity); err != nil {
// 		http.Error(w, "Invalid input", http.StatusBadRequest)
// 		return
// 	}

// 	activity.Username = claims.Username
// 	activity.Timestamp = time.Now() // Set the current timestamp
// 	activitiesCollection := client.Database("eventdb").Collection("activities")
// 	_, err = activitiesCollection.InsertOne(context.TODO(), activity)
// 	if err != nil {
// 		http.Error(w, "Failed to log activity", http.StatusInternalServerError)
// 		return
// 	}
// 	log.Println(activity)
// 	w.WriteHeader(http.StatusCreated) // 201 Created
// }

// // Fetch activity feed
// func getActivityFeed(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	tokenString := r.Header.Get("Authorization")
// 	claims := &Claims{}
// 	_, err := jwt.ParseWithClaims(tokenString[7:], claims, func(token *jwt.Token) (interface{}, error) {
// 		return jwtSecret, nil
// 	})
// 	if err != nil {
// 		http.Error(w, "Unauthorized", http.StatusUnauthorized)
// 		return
// 	}

// 	activitiesCollection := client.Database("eventdb").Collection("activities")
// 	cursor, err := activitiesCollection.Find(context.TODO(), bson.M{"username": claims.Username})
// 	if err != nil {
// 		http.Error(w, "Failed to fetch activities", http.StatusInternalServerError)
// 		return
// 	}
// 	defer cursor.Close(context.TODO())

// 	var activities []Activity
// 	if err := cursor.All(context.TODO(), &activities); err != nil {
// 		http.Error(w, "Failed to decode activities", http.StatusInternalServerError)
// 		return
// 	}
// 	log.Println(activities)
// 	json.NewEncoder(w).Encode(activities)
// }
