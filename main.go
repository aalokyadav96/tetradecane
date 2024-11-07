package main

import (
	"context"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	router := httprouter.New()
	router.GET("/", Index)
	router.GET("/activity", Index)
	router.GET("/about", Index)
	router.GET("/profile", Index)
	router.GET("/register", Index)
	router.GET("/login", Index)
	router.GET("/create", Index)
	router.GET("/place", Index)
	router.GET("/places", Index)
	router.GET("/events", Index)
	router.GET("/user/:username", Index)
	router.GET("/event/:eventid", Index)
	router.GET("/place/:placeid", Index)

	router.GET("/favicon.ico", Favicon)

	router.POST("/api/register", rateLimit(register))
	router.POST("/api/login", rateLimit(login))
	router.POST("/api/activity", authenticate(logActivity))
	router.GET("/api/profile", authenticate(getProfile))
	router.PUT("/api/profile", authenticate(editProfile))
	router.DELETE("/api/profile", authenticate(deleteProfile))
	router.POST("/api/follows/:id", authenticate(toggleFollow))
	// router.GET("/api/follows/:id", authenticate(doesFollow))
	router.GET("/api/followers", authenticate(getFollowers))
	router.GET("/api/following", authenticate(getFollowing))
	router.GET("/api/follow/suggestions", authenticate(suggestFollowers))
	router.GET("/api/activity", authenticate(getActivityFeed))
	router.GET("/api/user/:username", getUserProfile)

	router.GET("/api/events", getEvents)
	router.POST("/api/event", authenticate(createEvent))
	router.GET("/api/event/:eventid", getEvent)
	router.PUT("/api/event/:eventid", authenticate(editEvent))
	router.DELETE("/api/event/:eventid", authenticate(deleteEvent))

	router.POST("/api/event/:eventid/review", authenticate(addReview))

	router.POST("/api/event/:eventid/media", authenticate(addMedia))
	router.GET("/api/event/:eventid/media/:id", getMedia)
	router.GET("/api/event/:eventid/media", getMedias)
	router.DELETE("/api/event/:eventid/media/:id", authenticate(deleteMedia))

	router.POST("/api/event/:eventid/merch", authenticate(createMerch))
	router.GET("/api/event/:eventid/merch", getMerchs)
	router.GET("/api/event/:eventid/merch/:merchid", getMerch)
	router.PUT("/api/event/:eventid/merch/:merchid", authenticate(editMerch))
	router.DELETE("/api/event/:eventid/merch/:merchid", authenticate(deleteMerch))

	router.POST("/api/event/:eventid/ticket", authenticate(createTick))
	router.GET("/api/event/:eventid/ticket", getTicks)
	router.POST("/api/event/:eventid/ticket/:ticketid", buyTicket)
	router.PUT("/api/event/:eventid/ticket/:tickid", authenticate(editTick))
	router.DELETE("/api/event/:eventid/ticket/:tickid", authenticate(deleteTick))

	router.GET("/api/places", getPlaces)
	router.POST("/api/place", authenticate(createPlace))
	router.GET("/api/place/:placeid", getPlace)
	router.PUT("/api/place/:placeid", authenticate(editPlace))
	router.DELETE("/api/place/:placeid", authenticate(deletePlace))
	router.DELETE("/api/place/:placeid/review", authenticate(addReview))
	router.DELETE("/api/place/:placeid/media", authenticate(addMedia))
	router.POST("/api/place/:placeid/merch", authenticate(createMerch))
	router.GET("/api/place/:placeid/merch/:merchid", getMerch)
	router.PUT("/api/place/:placeid/merch/:merchid", authenticate(editMerch))
	router.DELETE("/api/place/:placeid/merch/:merchid", authenticate(deleteMerch))

	// // CORS setup
	// c := cors.New(cors.Options{
	// 	AllowedOrigins:   []string{"*"},
	// 	AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE"},
	// 	AllowCredentials: true,
	// })

	// Serve static files (HTML, CSS, JS)
	router.ServeFiles("/css/*filepath", http.Dir("css"))
	router.ServeFiles("/js/*filepath", http.Dir("js"))
	router.ServeFiles("/uploads/*filepath", http.Dir("uploads"))
	router.ServeFiles("/userpic/*filepath", http.Dir("userpic"))
	router.ServeFiles("/merchpic/*filepath", http.Dir("merchpic"))
	router.ServeFiles("/eventpic/*filepath", http.Dir("eventpic"))
	router.ServeFiles("/placepic/*filepath", http.Dir("placepic"))
	http.ListenAndServe("localhost:4000", router)
	// Initialize the HTTP server
	// server := &http.Server{
	// 	Addr:    ":4000",
	// 	Handler: c.Handler(router),
	// }

	// // Start server in a goroutine to handle graceful shutdown
	// go func() {
	// 	log.Println("Server started on port 4000")
	// 	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
	// 		log.Fatalf("Could not listen on port 4000: %v", err)
	// 	}
	// }()

	// // Graceful shutdown listener
	// shutdownChan := make(chan os.Signal, 1)
	// signal.Notify(shutdownChan, os.Interrupt, syscall.SIGTERM)

	// // Wait for termination signal
	// <-shutdownChan
	// log.Println("Shutting down gracefully...")

	// // Log active connections before shutdown
	// // log.Println("Active connections:", len(http.DefaultServeMux))

	// // Attempt to gracefully shut down the server
	// if err := server.Shutdown(context.Background()); err != nil {
	// 	log.Fatalf("Server shutdown failed: %v", err)
	// }
	// log.Println("Server stopped")
}

var (
	client         *mongo.Client
	userCollection *mongo.Collection
	jwtSecret      = []byte("your_secret_key") // Replace with your secret key
)

// Initialize MongoDB connection
func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	var err error
	client, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	userCollection = client.Database("eventdb").Collection("users")
}

// // Initialize MongoDB connection
// func init() {

// 	// Load environment variables from .env file
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatalf("Error loading .env file")
// 	}

// 	// Get the MongoDB URI from the environment variable
// 	mongoURI := os.Getenv("MONGODB_URI")
// 	if mongoURI == "" {
// 		log.Fatalf("MONGODB_URI environment variable is not set")
// 	}

// 	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
// 	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
// 	opts := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

// 	// Create a new client and connect to the server
// 	client, err := mongo.Connect(context.TODO(), opts)
// 	if err != nil {
// 		panic(err)
// 	}

// 	defer func() {
// 		if err = client.Disconnect(context.TODO()); err != nil {
// 			panic(err)
// 		}
// 	}()

// 	// Send a ping to confirm a successful connection
// 	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{"ping", 1}}).Err(); err != nil {
// 		panic(err)
// 	}
// 	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")
// 	userCollection = client.Database("your_database").Collection("users")
// }

// func init() {
// 	// Load environment variables from .env file
// 	err := godotenv.Load()
// 	if err != nil {
// 		log.Fatalf("Error loading .env file")
// 	}

// 	// Get the MongoDB URI from the environment variable
// 	mongoURI := os.Getenv("MONGODB_URI")
// 	if mongoURI == "" {
// 		log.Fatalf("MONGODB_URI environment variable is not set")
// 	}

// 	// Set up MongoDB client options
// 	clientOptions := options.Client().ApplyURI(mongoURI)

// 	// Connect to MongoDB
// 	client, err = mongo.Connect(context.TODO(), clientOptions)
// 	if err != nil {
// 		log.Fatalf("Failed to connect to MongoDB: %v", err)
// 	}

// 	// Set the user collection
// 	userCollection = client.Database("your_database").Collection("users")
// }

func Index(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	tmpl.ExecuteTemplate(w, "index.html", nil)
}

func Favicon(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// Serve the favicon if needed
}
