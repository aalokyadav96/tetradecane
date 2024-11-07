package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
)

func addMedia(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")

	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // Limit the size to 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve form values
	// Retrieve additional form values (if any)
	// mediaType := r.FormValue("type")
	// caption := r.FormValue("caption")
	// description := r.FormValue("description")

	// var media Media
	// Create a new Media instance
	media := Media{
		ID:      generateID(16),
		EventID: eventID,
		Type:    "image", // Use type from form or set default
		Caption: "Needs better forms",
		// Description: description,
	}

	// Handle banner file upload
	bannerFile, _, err := r.FormFile("media")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "Error retrieving banner file", http.StatusBadRequest)
		return
	}
	defer bannerFile.Close()

	if bannerFile != nil {
		// Save the banner image logic here
		out, err := os.Create("./uploads/" + media.ID + ".jpg")
		if err != nil {
			http.Error(w, "Error saving banner", http.StatusInternalServerError)
			return
		}
		defer out.Close()
		if _, err := io.Copy(out, bannerFile); err != nil {
			http.Error(w, "Error saving banner", http.StatusInternalServerError)
			return
		}
		media.URL = media.ID + ".jpg"
	}

	media.URL = media.ID + ".jpg"

	// Insert merch into MongoDB
	collection := client.Database("eventdb").Collection("media")
	_, err = collection.InsertOne(context.TODO(), media)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the created merchandise
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(media)
}

// // getMedia retrieves a specific media file
// func getMedia(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	eventID := ps.ByName("eventid")
// 	mediaID := ps.ByName("id")

// 	collection := client.Database("eventdb").Collection("media")
// 	var media Media
// 	err := collection.FindOne(context.TODO(), bson.M{"eventid": eventID, "id": mediaID}).Decode(&media)
// 	if err != nil {
// 		http.Error(w, "Media not found", http.StatusNotFound)
// 		return
// 	}

// 	// Serve the media file
// 	http.ServeFile(w, r, media.URL)
// }

// func getMedia(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
//     eventID := ps.ByName("eventid")
//     mediaID := ps.ByName("id")

//     // Debugging logs
//     fmt.Println("EventID:", eventID, "MediaID:", mediaID)

//     // Find the media document in the database
//     collection := client.Database("eventdb").Collection("media")
//     var media Media
//     err := collection.FindOne(context.TODO(), bson.M{"eventid": eventID, "id": mediaID}).Decode(&media)
//     if err != nil {
//         http.Error(w, "Media not found", http.StatusNotFound)
//         return
//     }

//     // Ensure the file URL is correct relative to your server's file path
//     filePath := "./uploads/" + media.URL
//     fmt.Println("Serving file:", filePath)

//     // Serve the file
//     http.ServeFile(w, r, filePath)
// }

// getMedia retrieves a specific media file for a given event and media ID
func getMedia(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid") // Extract the event ID from the URL parameters
	mediaID := ps.ByName("id")      // Extract the media ID from the URL parameters

	// Query the database for the media document using both eventID and mediaID
	collection := client.Database("eventdb").Collection("media")
	var media Media
	err := collection.FindOne(context.TODO(), bson.M{"eventid": eventID, "id": mediaID}).Decode(&media)
	if err != nil {
		// If there's an error (e.g., no matching media found), send a 404 response
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(media)
}

// getMedias retrieves all media for a specific event
func getMedias(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")

	collection := client.Database("eventdb").Collection("media")
	cursor, err := collection.Find(context.TODO(), bson.M{"eventid": eventID})
	if err != nil {
		http.Error(w, "Failed to retrieve media", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var medias []Media
	for cursor.Next(context.TODO()) {
		var media Media
		if err := cursor.Decode(&media); err != nil {
			http.Error(w, "Failed to decode media", http.StatusInternalServerError)
			return
		}
		medias = append(medias, media)
	}
	if len(medias) == 0 {
		medias = []Media{}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(medias)
}

// deleteMedia deletes a specific media file
func deleteMedia(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")
	mediaID := ps.ByName("id")

	collection := client.Database("eventdb").Collection("media")
	var media Media
	err := collection.FindOne(context.TODO(), bson.M{"eventid": eventID, "id": mediaID}).Decode(&media)
	if err != nil {
		http.Error(w, "Media not found", http.StatusNotFound)
		return
	}

	// Remove the media entry from the database
	_, err = collection.DeleteOne(context.TODO(), bson.M{"eventid": eventID, "id": mediaID})
	if err != nil {
		http.Error(w, "Failed to delete media from database", http.StatusInternalServerError)
		return
	}

	// Optionally, remove the file from the filesystem
	os.Remove(media.URL)

	// w.WriteHeader(http.StatusNoContent)
	sendResponse(w, http.StatusNoContent, map[string]string{"": ""}, "Delete successful", nil)
}
