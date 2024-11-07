package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

// Handle retrieving another user's profile
func getUserProfile(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	username := ps.ByName("username")
	var user User

	// Retrieve the user by username
	err := userCollection.FindOne(context.TODO(), bson.M{"username": username}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "User not found", http.StatusNotFound)
			log.Printf("User not found: %s", username)
			return
		}
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		log.Printf("Error retrieving user: %v", err)
		return
	}

	// Prepare the user profile response
	var userProfile UserProfileResponse
	userProfile.UserID = user.UserID
	userProfile.Username = user.Username
	userProfile.Email = user.Email
	userProfile.Bio = user.Bio
	userProfile.ProfilePicture = user.ProfilePicture
	userProfile.SocialLinks = user.SocialLinks

	// Get the ID of the requesting user from the context
	requestingUserID, ok := r.Context().Value(userIDKey).(string)
	if ok {
		log.Println("Requesting User ID:", requestingUserID)

		// Check if the requesting user is following the target user
		userProfile.IsFollowing = contains(user.Followers, requestingUserID)
	}

	log.Println("User Profile Response:", userProfile)

	// Send the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userProfile)
}

func editProfile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tokenString := r.Header.Get("Authorization")
	claims := &Claims{}

	// Validate JWT token
	_, err := jwt.ParseWithClaims(tokenString[7:], claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse the multipart form
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Prepare an update document
	update := bson.M{}

	// Retrieve and update fields from the form
	if username := r.FormValue("username"); username != "" {
		update["username"] = username
	}
	if email := r.FormValue("email"); email != "" {
		update["email"] = email
	}
	if bio := r.FormValue("bio"); bio != "" {
		update["bio"] = bio
	}
	if phoneNumber := r.FormValue("phone_number"); phoneNumber != "" {
		update["phone_number"] = phoneNumber
	}

	// Handle social links
	if socialLinks := r.FormValue("social_links"); socialLinks != "" {
		var links map[string]string
		if err := json.Unmarshal([]byte(socialLinks), &links); err == nil {
			update["social_links"] = links
		}
	}

	// Optional: handle password update
	if password := r.FormValue("password"); password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Failed to hash password", http.StatusInternalServerError)
			return
		}
		update["password"] = string(hashedPassword)
	}

	// Handle profile picture upload
	if file, _, err := r.FormFile("profile_picture"); err == nil {
		defer file.Close()

		// Save the file to a predefined location (adjust path and filename handling)
		out, err := os.Create("./userpic/" + claims.Username + ".jpg")
		if err != nil {
			log.Printf("Error creating file: %v", err)
			http.Error(w, "Failed to save profile picture", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		// Copy the uploaded file to the destination
		if _, err := io.Copy(out, file); err != nil {
			http.Error(w, "Failed to save profile picture", http.StatusInternalServerError)
			return
		}

		// Update the profile picture field in the update document
		update["profile_picture"] = "/" + claims.Username + ".jpg"
	}

	// Update the user in the database
	_, err = userCollection.UpdateOne(context.TODO(), bson.M{"username": claims.Username}, bson.M{"$set": update})
	if err != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent) // Send 204 No Content
}

// Handle profile retrieval
func getProfile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tokenString := r.Header.Get("Authorization")
	claims := &Claims{}
	jwt.ParseWithClaims(tokenString[7:], claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	var user User
	err := userCollection.FindOne(context.TODO(), bson.M{"username": claims.Username}).Decode(&user)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		log.Printf("User not found: %s", claims.Username)
		return
	}

	user.Password = "" // Do not return the password
	json.NewEncoder(w).Encode(user)
}

// Handle deleting the user's profile
func deleteProfile(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	userID := r.Context().Value(userIDKey).(string) // Get the user ID from context
	log.Println("beep: ", userID)
	_, err := userCollection.DeleteOne(context.TODO(), bson.M{"userid": userID})
	if err != nil {
		http.Error(w, "Error deleting profile", http.StatusInternalServerError)
		log.Printf("Error deleting user profile: %v", err)
		return
	}

	// w.WriteHeader(http.StatusNoContent) // No Content response
	log.Printf("User profile deleted: %s", userID)

	sendResponse(w, http.StatusOK, map[string]string{"": ""}, "Deletion successful", nil)
}
