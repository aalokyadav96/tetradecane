package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

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

	// Prepare a map for updating fields
	updateFields := bson.M{}

	// Only set the fields that are provided in the form
	if title := r.FormValue("title"); title != "" {
		updateFields["title"] = title
	}

	if date := r.FormValue("date"); date != "" {
		updateFields["date"] = date
	}

	if place := r.FormValue("place"); place != "" {
		updateFields["place"] = place
	}

	if location := r.FormValue("location"); location != "" {
		updateFields["location"] = location
	}

	if description := r.FormValue("description"); description != "" {
		updateFields["description"] = description
	}

	// Validate required fields
	if updateFields["title"] == "" || updateFields["location"] == "" || updateFields["description"] == "" {
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

	// If a new banner is uploaded, save it and update the field
	if bannerFile != nil {
		// Ensure the directory exists
		if err := os.MkdirAll("./eventpic", os.ModePerm); err != nil {
			http.Error(w, "Error creating directory for banner", http.StatusInternalServerError)
			return
		}

		// Save the banner image
		out, err := os.Create("./eventpic/" + eventID + ".jpg")
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

		// Update the banner image path in the updateFields map
		updateFields["banner_image"] = eventID + ".jpg"
	}

	// Update the event in MongoDB (only the fields that have changed)
	collection := client.Database("eventdb").Collection("events")
	updateFields["updated_at"] = time.Now() // Update the timestamp for the update

	// Perform the update query
	_, err = collection.UpdateOne(
		context.TODO(),
		bson.M{"eventid": eventID},
		bson.M{"$set": updateFields},
	)
	if err != nil {
		http.Error(w, "Error updating event", http.StatusInternalServerError)
		return
	}

	// Respond with the updated fields
	w.WriteHeader(http.StatusOK) // 200 OK
	updatedEvent := Event{EventID: eventID}
	for key, value := range updateFields {
		// Assign updated values back to the Event struct
		switch key {
		case "title":
			updatedEvent.Title = value.(string)
		case "description":
			updatedEvent.Description = value.(string)
		case "place":
			updatedEvent.Place = value.(string)
		case "date":
			updatedEvent.Date = value.(string)
		case "location":
			updatedEvent.Location = value.(string)
			// case "banner_image":
			// 	updatedEvent.BannerImage = value.(string)
		}
	}

	// Send the updated event as the response
	if err := json.NewEncoder(w).Encode(updatedEvent); err != nil {
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
