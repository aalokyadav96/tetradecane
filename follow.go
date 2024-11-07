package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
)

// Handle retrieving followers
func getFollowers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	followers := []User{}
	for _, followerID := range user.Follows {
		var follower User
		if err := userCollection.FindOne(context.TODO(), bson.M{"userid": followerID}).Decode(&follower); err == nil {
			followers = append(followers, follower)
		}
	}

	json.NewEncoder(w).Encode(followers)
}

// Handle retrieving following
func getFollowing(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	following := []User{}
	for _, followingID := range user.Follows {
		var followUser User
		if err := userCollection.FindOne(context.TODO(), bson.M{"userid": followingID}).Decode(&followUser); err == nil {
			following = append(following, followUser)
		}
	}

	json.NewEncoder(w).Encode(following)
}

// Handle suggesting users to follow
func suggestFollowers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	// Suggest users excluding the current user and already followed users
	suggestedUsers := []User{}
	cursor, err := userCollection.Find(context.TODO(), bson.M{"username": bson.M{"$ne": user.Username}})
	if err != nil {
		http.Error(w, "Failed to fetch suggestions", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {
		var suggestedUser User
		if err := cursor.Decode(&suggestedUser); err == nil && !contains(user.Follows, suggestedUser.Username) {
			suggestedUser.Password = ""
			suggestedUsers = append(suggestedUsers, suggestedUser)
		}
	}

	json.NewEncoder(w).Encode(suggestedUsers)
}

// Toggle Follow function
func toggleFollow(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	userId, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	followedUserId := ps.ByName("id")
	if followedUserId == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	log.Printf("User %s is trying to toggle follow for user %s", userId, followedUserId)

	// Retrieve the current user
	var currentUser User
	err := userCollection.FindOne(context.TODO(), bson.M{"userid": userId}).Decode(&currentUser)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Check if the user is already following the followed user
	isFollowing := false
	for _, followedID := range currentUser.Follows {
		if followedID == followedUserId {
			isFollowing = true
			break
		}
	}

	if isFollowing {
		// Unfollow: remove followedUserId from currentUser.Follows
		currentUser.Follows = removeString(currentUser.Follows, followedUserId)

		// Remove currentUser.UserID from followed user's Followers
		_, err = userCollection.UpdateOne(context.TODO(), bson.M{"userid": followedUserId}, bson.M{
			"$pull": bson.M{"followers": userId},
		})
		if err != nil {
			log.Printf("Error updating followers: %v", err)
			http.Error(w, "Failed to update followers", http.StatusInternalServerError)
			return
		}
	} else {
		// Follow: add followedUserId to currentUser.Follows
		currentUser.Follows = append(currentUser.Follows, followedUserId)

		// Add currentUser.UserID to followed user's Followers
		_, err = userCollection.UpdateOne(context.TODO(), bson.M{"userid": followedUserId}, bson.M{
			"$addToSet": bson.M{"followers": userId},
		})
		if err != nil {
			log.Printf("Error updating followers: %v", err)
			http.Error(w, "Failed to update followers", http.StatusInternalServerError)
			return
		}
	}

	// Update the current user's follows array
	_, err = userCollection.UpdateOne(context.TODO(), bson.M{"userid": userId}, bson.M{
		"$set": bson.M{"follows": currentUser.Follows},
	})
	if err != nil {
		log.Printf("Error updating follows: %v", err)
		http.Error(w, "Failed to update follows", http.StatusInternalServerError)
		return
	}

	// Return the updated follow status in the response
	response := map[string]bool{"isFollowing": !isFollowing} // Toggle status
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
