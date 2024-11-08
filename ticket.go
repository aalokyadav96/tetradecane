package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/bson"
)

// Create Ticket
func createTick(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")

	// Retrieve form values
	name := r.FormValue("name")
	priceStr := r.FormValue("price")
	log.Println("\n\n\n\npriceStr : \n\n ", priceStr)
	// Convert the string to float64
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		// Handle error (e.g., invalid input)
		http.Error(w, "Invalid price value", http.StatusBadRequest)
		return
	}
	quantity, err := strconv.Atoi(r.FormValue("quantity"))
	if err != nil {
		http.Error(w, "Invalid quantity value", http.StatusBadRequest)
		return
	}

	// Create a new Ticket instance
	tick := Ticket{
		EventID:  eventID,
		Name:     name,
		Price:    price,
		Quantity: quantity,
	}

	tick.TicketID = generateID(12)

	// Insert ticket into MongoDB
	collection := client.Database("eventdb").Collection("ticks")
	_, err = collection.InsertOne(context.TODO(), tick)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the created ticket
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tick)
}

// Get all Tickets for an Event
func getTicks(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")

	collection := client.Database("eventdb").Collection("ticks")

	var tickList []Ticket
	filter := bson.M{"eventid": eventID}

	// Query the database
	cursor, err := collection.Find(context.Background(), filter)
	if err != nil {
		http.Error(w, "Failed to fetch tickets", http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	// Iterate through the cursor and decode each document into the ticketList
	for cursor.Next(context.Background()) {
		var tick Ticket
		if err := cursor.Decode(&tick); err != nil {
			http.Error(w, "Failed to decode ticket", http.StatusInternalServerError)
			return
		}
		tickList = append(tickList, tick)
	}

	// Check for cursor errors
	if err := cursor.Err(); err != nil {
		http.Error(w, "Cursor error", http.StatusInternalServerError)
		return
	}
	if len(tickList) == 0 {
		tickList = []Ticket{}
	}

	// Respond with the ticket data
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(tickList); err != nil {
		http.Error(w, "Failed to encode ticket data", http.StatusInternalServerError)
	}
}

// Edit Ticket
func editTick(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")
	tickID := ps.ByName("ticketid")
	var tick Ticket
	json.NewDecoder(r.Body).Decode(&tick)

	// Update the ticket in MongoDB
	collection := client.Database("eventdb").Collection("ticks")
	_, err := collection.UpdateOne(context.TODO(), bson.M{"eventid": eventID, "ticketid": tickID}, bson.M{"$set": tick})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tick)
}

// Delete Ticket
func deleteTick(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")
	tickID := ps.ByName("ticketid")

	// Delete the ticket from MongoDB
	collection := client.Database("eventdb").Collection("ticks")
	_, err := collection.DeleteOne(context.TODO(), bson.M{"eventid": eventID, "ticketid": tickID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// w.WriteHeader(http.StatusNoContent)
	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Ticket deleted successfully",
	})
}

// Buy Ticket
func buyTicket(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	eventID := ps.ByName("eventid")
	ticketID := ps.ByName("ticketid")

	// Find the ticket in the database
	collection := client.Database("eventdb").Collection("ticks")
	var ticket Ticket // Define the Ticket struct based on your schema
	err := collection.FindOne(context.TODO(), bson.M{"eventid": eventID, "ticketid": ticketID}).Decode(&ticket)
	if err != nil {
		http.Error(w, "Ticket not found or other error", http.StatusNotFound)
		return
	}

	// Check if there are tickets available
	if ticket.Quantity <= 0 {
		http.Error(w, "No tickets available for purchase", http.StatusBadRequest)
		return
	}

	// Decrease the ticket quantity
	update := bson.M{"$inc": bson.M{"quantity": -1}}
	_, err = collection.UpdateOne(context.TODO(), bson.M{"eventid": eventID, "ticketid": ticketID}, update)
	if err != nil {
		http.Error(w, "Failed to update ticket quantity", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Ticket purchased successfully",
	})
}
