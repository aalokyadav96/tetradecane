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
	"golang.org/x/crypto/bcrypt"
)

// JWT claims
type Claims struct {
	Username string `json:"username"`
	UserID   string `json:"userId"`
	jwt.RegisteredClaims
}

// Updated Login Function
// func login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
// 	var user User
// 	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
// 		http.Error(w, "Invalid input", http.StatusBadRequest)
// 		return
// 	}

// 	var storedUser User
// 	err := userCollection.FindOne(context.TODO(), bson.M{"username": user.Username}).Decode(&storedUser)
// 	if err != nil || bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)) != nil {
// 		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
// 		return
// 	}

// 	claims := &Claims{
// 		Username: storedUser.Username,
// 		UserID:   storedUser.UserID, // Make sure to set UserID here
// 		RegisteredClaims: jwt.RegisteredClaims{
// 			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 72)),
// 		},
// 	}
// 	log.Print(storedUser.UserID)
// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
// 	tokenString, err := token.SignedString(jwtSecret)
// 	if err != nil {
// 		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
// 		return
// 	}
// 	sendResponse(w, http.StatusOK, map[string]string{"token": tokenString,"userid": storedUser.UserID}, "Login successful", nil)
// }

func login(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	var storedUser User
	err := userCollection.FindOne(context.TODO(), bson.M{"username": user.Username}).Decode(&storedUser)
	if err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(storedUser.Password), []byte(user.Password)); err != nil {
		http.Error(w, "Invalid username or password", http.StatusUnauthorized)
		return
	}

	// Create JWT claims
	claims := &Claims{
		Username: storedUser.Username,
		UserID:   storedUser.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret) // Ensure jwtSecret is a byte array
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Send response
	sendResponse(w, http.StatusOK, map[string]string{"token": tokenString, "userid": storedUser.UserID}, "Login successful", nil)
}

// Handle user registration
func register(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}
	log.Printf("Registering user: %s", user.Username)

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Failed to hash password for user %s: %v", user.Username, err)
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}
	user.Password = string(hashedPassword)
	user.UserID = "u" + GenerateName(10)
	_, err = userCollection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Printf("User already exists: %s", user.Username)
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	response := map[string]interface{}{
		"status":  http.StatusCreated,
		"message": "",
		"data":    "",
	}
	json.NewEncoder(w).Encode(response)
	// w.WriteHeader(http.StatusCreated)
}

type contextKey string

const userIDKey contextKey = "userId"

// Authenticate middleware
func authenticate(next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}

		if len(tokenString) < 7 || tokenString[:7] != "Bearer " {
			http.Error(w, "Invalid token format", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString[7:], claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Store UserID in context
		ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
		next(w, r.WithContext(ctx), ps) // Call the next handler with new context
	}
}
