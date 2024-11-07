package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func createEvent(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Parse the multipart form with a 10MB limit
	if err := r.ParseMultipartForm(10 << 20); err != nil { // Limit upload size to 10MB
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	var event Event
	// Get the event data from the form (assuming it's passed as JSON string)
	err := json.Unmarshal([]byte(r.FormValue("event")), &event)
	if err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Retrieve the ID of the requesting user from the context
	requestingUserID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}
	event.CreatorID = requestingUserID

	// Generate a unique EventID
	event.EventID = generateID(14)

	// Handle the banner image upload (if present)
	bannerFile, _, err := r.FormFile("banner")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "Error retrieving banner file", http.StatusBadRequest)
		return
	}

	// If a banner file is provided, process it
	if bannerFile != nil {
		// Ensure the directory exists
		if err := os.MkdirAll("./eventpic", os.ModePerm); err != nil {
			http.Error(w, "Error creating directory for banner", http.StatusInternalServerError)
			return
		}

		// Save the banner image
		out, err := os.Create("./eventpic/" + event.EventID + ".jpg")
		if err != nil {
			http.Error(w, "Error saving banner", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		// Copy the content from the uploaded file to the destination file
		if _, err := io.Copy(out, bannerFile); err != nil {
			http.Error(w, "Error saving banner", http.StatusInternalServerError)
			return
		}

		// Set the event's banner image field with the saved image path
		event.BannerImage = event.EventID + ".jpg"
	}

	// Insert the event into MongoDB
	collection := client.Database("eventdb").Collection("events")
	_, err = collection.InsertOne(context.TODO(), event)
	if err != nil {
		http.Error(w, "Error saving event", http.StatusInternalServerError)
		return
	}

	// Respond with the created event
	w.WriteHeader(http.StatusCreated) // 201 Created
	if err := json.NewEncoder(w).Encode(event); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func getEvents(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Set the response header to indicate JSON content type
	w.Header().Set("Content-Type", "application/json")

	collection := client.Database("eventdb").Collection("events")

	// Find all events
	cursor, err := collection.Find(context.TODO(), bson.M{})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.TODO())

	var events []Event
	if err = cursor.All(context.TODO(), &events); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Encode the list of events as JSON and write to the response
	json.NewEncoder(w).Encode(events)
}

// func getEvent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	id := ps.ByName("eventid")

// 	collection := client.Database("eventdb").Collection("events")
// 	var event Event
// 	err := collection.FindOne(context.TODO(), bson.M{"eventid": id}).Decode(&event)
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusNotFound)
// 		return
// 	}
// 	if event.Tickets == nil {
// 		event.Tickets = []Ticket{} // Initialize as an empty array if it's nil
// 	}

// 	if event.Media == nil {
// 		event.Media = []Media{} // Initialize as an empty array if it's nil
// 	}

// 	if event.Merch == nil {
// 		event.Merch = []Merch{} // Initialize as an empty array if it's nil
// 	}

// 	json.NewEncoder(w).Encode(event)
// }

func getEvent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id := ps.ByName("eventid")

	// Fetch event data from the "events" collection
	eventsCollection := client.Database("eventdb").Collection("events")
	var event Event
	err := eventsCollection.FindOne(context.TODO(), bson.M{"eventid": id}).Decode(&event)
	if err != nil {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Initialize fields as empty slices if they're nil
	if event.Tickets == nil {
		event.Tickets = []Ticket{}
	}
	if event.Media == nil {
		event.Media = []Media{}
	}
	if event.Merch == nil {
		event.Merch = []Merch{}
	}

	// Fetch tickets data
	ticketsCollection := client.Database("eventdb").Collection("ticks")
	ticketsCursor, err := ticketsCollection.Find(context.TODO(), bson.M{"eventid": id})
	if err == nil {
		defer ticketsCursor.Close(context.TODO())
		for ticketsCursor.Next(context.TODO()) {
			var ticket Ticket
			if err := ticketsCursor.Decode(&ticket); err == nil {
				event.Tickets = append(event.Tickets, ticket)
			}
		}
	}

	// Fetch media data
	mediaCollection := client.Database("eventdb").Collection("media")
	mediaCursor, err := mediaCollection.Find(context.TODO(), bson.M{"eventid": id})
	if err == nil {
		defer mediaCursor.Close(context.TODO())
		for mediaCursor.Next(context.TODO()) {
			var media Media
			if err := mediaCursor.Decode(&media); err == nil {
				event.Media = append(event.Media, media)
			}
		}
	}

	// Fetch merch data
	merchCollection := client.Database("eventdb").Collection("merch")
	merchCursor, err := merchCollection.Find(context.TODO(), bson.M{"eventid": id})
	if err == nil {
		defer merchCursor.Close(context.TODO())
		for merchCursor.Next(context.TODO()) {
			var merch Merch
			if err := merchCursor.Decode(&merch); err == nil {
				event.Merch = append(event.Merch, merch)
			}
		}
	}

	// Send the combined event data with tickets, media, and merch
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(event); err != nil {
		http.Error(w, "Failed to encode event data", http.StatusInternalServerError)
	}
}

func editEvent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")

	// Parse the multipart form with a 10MB limit
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10 MB limit
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// Retrieve the event data from the form
	var event Event
	event.Title = r.FormValue("title")
	event.Date = r.FormValue("date")
	event.Place = r.FormValue("place")
	event.Location = r.FormValue("location")
	event.Description = r.FormValue("description")
	event.EventID = eventID // Ensure we set the ID

	// Validate required fields
	if event.Title == "" || event.Location == "" || event.Description == "" {
		http.Error(w, "Title, Location, and Description are required", http.StatusBadRequest)
		return
	}

	// Handle banner file upload if present
	bannerFile, _, err := r.FormFile("event-banner")
	if err != nil && err != http.ErrMissingFile {
		http.Error(w, "Error retrieving banner file", http.StatusBadRequest)
		return
	}

	// Close the bannerFile if it was opened
	defer func() {
		if bannerFile != nil {
			bannerFile.Close()
		}
	}()

	// If no banner file is uploaded, retain the existing banner image (if any)
	if bannerFile == nil {
		// If no new banner is uploaded, retain the existing banner image (no change)
		log.Print("No new banner uploaded, using existing image")
	} else {
		// Ensure the directory exists
		if err := os.MkdirAll("./eventpic", os.ModePerm); err != nil {
			http.Error(w, "Error creating directory for banner", http.StatusInternalServerError)
			return
		}

		// Save the banner image
		out, err := os.Create("./eventpic/" + event.EventID + ".jpg")
		if err != nil {
			http.Error(w, "Error saving banner", http.StatusInternalServerError)
			return
		}
		defer out.Close()

		// Copy the content of the uploaded file to the destination file
		if _, err := io.Copy(out, bannerFile); err != nil {
			http.Error(w, "Error saving banner", http.StatusInternalServerError)
			return
		}

		// Update the banner image path
		event.BannerImage = event.EventID + ".jpg"
	}

	// Update the event in MongoDB
	collection := client.Database("eventdb").Collection("events")
	_, err = collection.UpdateOne(context.TODO(), bson.M{"eventid": eventID}, bson.M{"$set": event})
	if err != nil {
		http.Error(w, "Error updating event", http.StatusInternalServerError)
		return
	}

	// Respond with the updated event
	w.WriteHeader(http.StatusOK) // 200 OK
	if err := json.NewEncoder(w).Encode(event); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func deleteEvent(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")

	// Get the ID of the requesting user from the context
	requestingUserID, ok := r.Context().Value(userIDKey).(string)
	if !ok {
		http.Error(w, "Invalid user", http.StatusBadRequest)
		return
	}

	// Get the event details to verify the creator
	collection := client.Database("eventdb").Collection("events")
	var event Event
	err := collection.FindOne(context.TODO(), bson.M{"eventid": eventID}).Decode(&event)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			http.Error(w, "Event not found", http.StatusNotFound)
		} else {
			http.Error(w, "Error retrieving event", http.StatusInternalServerError)
		}
		return
	}

	// Check if the requesting user is the creator of the event
	if event.CreatorID != requestingUserID {
		http.Error(w, "Unauthorized to delete this event", http.StatusForbidden)
		return
	}

	// Delete the event from MongoDB
	result, err := collection.DeleteOne(context.TODO(), bson.M{"eventid": eventID})
	if err != nil {
		http.Error(w, "Error deleting event", http.StatusInternalServerError)
		return
	}

	// Check if the event was found and deleted
	if result.DeletedCount == 0 {
		http.Error(w, "Event not found", http.StatusNotFound)
		return
	}

	// Send success response
	w.WriteHeader(http.StatusOK) // 200 OK
	response := map[string]string{"message": "Event deleted successfully"}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func addReview(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")
	var review Review
	json.NewDecoder(r.Body).Decode(&review)

	// Add review to MongoDB
	collection := client.Database("eventdb").Collection("events")
	_, err := collection.UpdateOne(context.TODO(), bson.M{"reviewid": eventID}, bson.M{"$push": bson.M{"reviews": review}})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// func addMedia(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
// 	eventID := ps.ByName("eventid")
// 	var media Media
// 	json.NewDecoder(r.Body).Decode(&media)

// 	// Add media to MongoDB
// 	collection := client.Database("eventdb").Collection("events")
// 	_, err := collection.UpdateOne(context.TODO(), bson.M{"eventid": eventID}, bson.M{"$push": bson.M{"media": media}})
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.WriteHeader(http.StatusNoContent)
// }
