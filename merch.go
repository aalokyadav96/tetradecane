package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
)

func createMerch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")

	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // Limit the size to 10 MB
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve form values
	name := r.FormValue("name")
	price, err := strconv.ParseFloat(r.FormValue("price"), 64)
	if err != nil {
		http.Error(w, "Invalid price value", http.StatusBadRequest)
		return
	}
	quantity, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		http.Error(w, "Invalid quantity value", http.StatusBadRequest)
		return
	}

	// Create a new Merch instance
	merch := Merch{
		EventID: eventID,
		Name:    name,
		Price:   price,
		Stock:   quantity,
	}

	merch.MerchID = generateID(14)

	// Handle banner file upload
	bannerFile, _, err := r.FormFile("image")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "Error retrieving banner file", http.StatusBadRequest)
		return
	}
	defer bannerFile.Close()

	if bannerFile != nil {
		// Save the banner image logic here
		out, err := os.Create("./merchpic/" + merch.MerchID + ".jpg")
		if err != nil {
			http.Error(w, "Error saving banner", http.StatusInternalServerError)
			return
		}
		defer out.Close()
		if _, err := io.Copy(out, bannerFile); err != nil {
			http.Error(w, "Error saving banner", http.StatusInternalServerError)
			return
		}
		merch.MerchPhoto = merch.MerchID + ".jpg"
	}

	// Insert merch into MongoDB
	collection := client.Database("eventdb").Collection("merch")
	_, err = collection.InsertOne(context.TODO(), merch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the created merchandise
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(merch)
}

func getMerch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")
	merchID := ps.ByName("merchid")

	collection := client.Database("eventdb").Collection("merch")
	var merch Merch
	err := collection.FindOne(context.TODO(), bson.M{"eventid": eventID, "merchid": merchID}).Decode(&merch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(merch)
}

func getMerchs(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")

	collection := client.Database("eventdb").Collection("merch")

	var merchList []Merch                // Use your Merch struct here
	filter := bson.M{"eventid": eventID} // Ensure this matches your BSON field name

	// Query the database
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		http.Error(w, "Failed to fetch merchandise", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	// Iterate through the cursor and decode each document into the merchList
	for cursor.Next(context.Background()) {
		var merch Merch
		if err := cursor.Decode(&merch); err != nil {
			http.Error(w, "Failed to decode merchandise", http.StatusInternalServerError)
			return
		}
		merchList = append(merchList, merch)
	}

	// Check for cursor errors
	if err := cursor.Err(); err != nil {
		http.Error(w, "Cursor error", http.StatusInternalServerError)
		return
	}
	if len(merchList) == 0 {
		merchList = []Merch{}
	}
	// Respond with the merchandise data
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(merchList); err != nil {
		http.Error(w, "Failed to encode merchandise data", http.StatusInternalServerError)
	}
}

func editMerch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")
	merchID := ps.ByName("merchid")
	var merch Merch
	json.NewDecoder(r.Body).Decode(&merch)

	// Update the merch in MongoDB
	collection := client.Database("eventdb").Collection("merch")
	_, err := collection.UpdateOne(context.TODO(), bson.M{"eventid": eventID, "merchid": merchID}, bson.M{"$set": merch})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(merch)
}

func deleteMerch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")
	merchID := ps.ByName("merchid")

	// Delete the merch from MongoDB
	collection := client.Database("eventdb").Collection("merch")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"eventid": eventID, "merchid": merchID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// w.WriteHeader(http.StatusNoContent)
	sendResponse(w, http.StatusNoContent, map[string]string{"": ""}, "Delete successful", nil)
}

// Buy Merch
func buyMerch(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")
	merchID := ps.ByName("merchid")

	// Find the merch in the database
	collection := client.Database("eventdb").Collection("merch")
	var merch Merch // Define the Merch struct based on your schema
	err := collection.FindOne(context.TODO(), bson.M{"eventid": eventID, "merchid": merchID}).Decode(&merch)
	if err != nil {
		http.Error(w, "Merch not found or other error", http.StatusNotFound)
		return
	}

	// Check if there are merchs available
	if merch.Stock <= 0 {
		http.Error(w, "No merchs available for purchase", http.StatusBadRequest)
		return
	}

	// Decrease the merch quantity
	update := bson.M{"$inc": bson.M{"stock": -1}}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"eventid": eventID, "merchid": merchID}, update)
	if err != nil {
		http.Error(w, "Failed to update merch quantity", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Merch purchased successfully",
	})
}
