package main

import (
	"time"
)

type User struct {
	ID          string    `json:"-" bson:"_id,omitempty"`
	UserID      string    `json:"userid" bson:"userid"`
	Username    string    `json:"username" bson:"username"`
	Email       string    `json:"email" bson:"email"`
	Password    string    `json:"-" bson:"password"`
	Role        string    `json:"role" bson:"role"`
	Name        string    `json:"name,omitempty" bson:"name,omitempty"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at"`
	PhoneNumber string    `json:"phone_number,omitempty" bson:"phone_number,omitempty"`
	Bio         string    `json:"bio,omitempty" bson:"bio,omitempty"`

	Preferences Preferences `json:"preferences,omitempty" bson:"preferences,omitempty"`

	IsActive       bool              `json:"is_active" bson:"is_active"`
	LastLogin      time.Time         `json:"last_login,omitempty" bson:"last_login,omitempty"`
	ProfilePicture string            `json:"profile_picture" bson:"profile_picture"`
	ProfileViews   int               `json:"profile_views,omitempty" bson:"profile_views,omitempty"`
	Address        string            `json:"address,omitempty" bson:"address,omitempty"`
	DateOfBirth    time.Time         `json:"date_of_birth,omitempty" bson:"date_of_birth,omitempty"`
	SocialLinks    map[string]string `json:"social_links,omitempty" bson:"social_links,omitempty"`
	IsVerified     bool              `json:"is_verified" bson:"is_verified"`
	Follows        []string          `json:"follows,omitempty" bson:"follows,omitempty"`
	Followers      []string          `json:"followers,omitempty" bson:"followers,omitempty"`
}

type Preferences struct {
	Theme             string `json:"theme" bson:"theme"`
	NotificationEmail bool   `json:"notification_email" bson:"notification_email"`
	// other preference fields
}

type Merch struct {
	MerchID    string  `json:"merchid" bson:"merchid"`
	EventID    string  `json:"eventid" bson:"eventid"` // Reference to Event ID
	Name       string  `json:"name" bson:"name"`
	Price      float64 `json:"price" bson:"price"`
	Stock      int     `json:"stock" bson:"stock"` // Number of items available
	MerchPhoto string  `json:"merch_pic" bson:"merch_pic"`
}

type Event struct {
	EventID     string `json:"eventid" bson:"eventid"`
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
	Place       string `json:"place" bson:"place"`
	Date        string `json:"date" bson:"date"`
	// Date        time.Time `json:"date" bson:"date"`

	Location string `json:"location" bson:"location"`
	// Address   string `json:"address" bson:"address"`
	CreatorID string `json:"creatorid" bson:"creatorid"`

	OrganizerName    string `json:"organizer_name" bson:"organizer_name"`
	OrganizerContact string `json:"organizer_contact" bson:"organizer_contact"`

	Tickets []Ticket `json:"tickets" bson:"tickets"`
	Media   []Media  `json:"media" bson:"media"`
	Merch   []Merch  `json:"merch" bson:"merch"`

	StartDateTime time.Time `json:"start_date_time" bson:"start_date_time"`
	EndDateTime   time.Time `json:"end_date_time" bson:"end_date_time"`

	Category          string `json:"category" bson:"category"`
	BannerImage       string `json:"banner_image" bson:"banner_image"`
	WebsiteURL        string `json:"website_url" bson:"website_url"`
	Status            string `json:"status" bson:"status"`
	AccessibilityInfo string `json:"accessibility_info" bson:"accessibility_info"`

	Reviews          []Review               `json:"reviews" bson:"reviews"`
	SocialMediaLinks []string               `json:"social_media_links" bson:"social_media_links"`
	Tags             []string               `json:"tags" bson:"tags"`
	CustomFields     map[string]interface{} `json:"custom_fields" bson:"custom_fields"`

	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type Place struct {
	PlaceID     string `json:"placeid" bson:"placeid"`
	Name        string `json:"name,omitempty" bson:"name,omitempty"`
	Description string `json:"description,omitempty" bson:"description,omitempty"`
	Banner      string `json:"banner,omitempty" bson:"banner,omitempty"`
	Address     string `json:"address,omitempty" bson:"address,omitempty"`
	City        string `json:"city,omitempty" bson:"city,omitempty"`
	// Location    string      `json:"location,omitempty" bson:"location,omitempty"`
	Country        string            `json:"country,omitempty" bson:"country,omitempty"`
	ZipCode        string            `json:"zipCode,omitempty" bson:"zipCode,omitempty"`
	Coordinates    Coordinates       `json:"coordinates,omitempty" bson:"coordinates,omitempty"`
	Capacity       int               `json:"capacity" bson:"capacity"`
	Phone          string            `json:"phone,omitempty" bson:"phone,omitempty"`
	Website        string            `json:"website,omitempty" bson:"website,omitempty"`
	Category       Category          `json:"category,omitempty" bson:"category,omitempty"`
	IsOpen         bool              `json:"isopen,omitempty" bson:"isopen,omitempty"`
	Distance       float64           `json:"distance,omitempty" bson:"distance,omitempty"`
	Status         PlaceStatus       `json:"status,omitempty" bson:"status,omitempty"`
	Views          int               `json:"views,omitempty" bson:"views,omitempty"`
	ReviewCount    int               `json:"reviewCount,omitempty" bson:"reviewCount,omitempty"`
	SocialLinks    map[string]string `json:"socialLinks,omitempty" bson:"socialLinks,omitempty"`
	CreatedBy      string            `json:"createdBy,omitempty" bson:"createdBy,omitempty"`
	UpdatedBy      string            `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
	CreatedAt      time.Time         `json:"created,omitempty" bson:"created,omitempty"`
	UpdatedAt      time.Time         `json:"updated,omitempty" bson:"updated,omitempty"`
	DeletedAt      *time.Time        `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
	Reviews        []Review          `json:"reviews,omitempty" bson:"reviews,omitempty"`
	Merch          []Merch           `json:"merch,omitempty" bson:"merch,omitempty"`
	Amenities      []string          `json:"amenities,omitempty" bson:"amenities,omitempty"`
	Events         []Event           `json:"events,omitempty" bson:"events,omitempty"`
	Tags           []string          `json:"tags,omitempty" bson:"tags,omitempty"`
	Medias         []Media           `json:"media,omitempty" bson:"media,omitempty"`
	OperatingHours []string          `json:"operatinghours,omitempty" bson:"operatinghours,omitempty"`
	Keywords       []string          `json:"keywords,omitempty" bson:"keywords,omitempty"`
}

type PlaceStatus string

const (
	Active   PlaceStatus = "active"
	Inactive PlaceStatus = "inactive"
	Closed   PlaceStatus = "closed"
)

type Activity struct {
	Username    string    `json:"username" bson:"username"`
	PlaceID     string    `json:"placeId,omitempty" bson:"placeId,omitempty"`
	Action      string    `json:"action,omitempty" bson:"action,omitempty"`
	PerformedBy string    `json:"performedBy,omitempty" bson:"performedBy,omitempty"`
	Timestamp   time.Time `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Details     string    `json:"details,omitempty" bson:"details,omitempty"`
	IPAddress   string    `json:"ipAddress,omitempty" bson:"ipAddress,omitempty"`
	DeviceInfo  string    `json:"deviceInfo,omitempty" bson:"deviceInfo,omitempty"`
}

// UserProfileResponse defines the structure for the user profile response
type UserProfileResponse struct {
	UserID         string            `json:"userid" bson:"userid"`
	Username       string            `json:"username" bson:"username"`
	Email          string            `json:"email" bson:"email"`
	Bio            string            `json:"bio,omitempty" bson:"bio,omitempty"`
	PhoneNumber    string            `json:"phone_number,omitempty" bson:"phone_number,omitempty"`
	ProfilePicture string            `json:"profile_picture" bson:"profile_picture"`
	IsFollowing    bool              `json:"is_following" bson:"is_following"` // Added here
	SocialLinks    map[string]string `json:"social_links,omitempty" bson:"social_links,omitempty"`
}

type Response struct {
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type Review struct {
	ReviewID    string    `json:"reviewid" bson:"reviewid"`
	EventID     string    `json:"eventid" bson:"eventid"` // Reference to Event ID
	UserID      string    `json:"userid" bson:"userid"`   // Reference to User ID
	Rating      int       `json:"rating" bson:"rating"`   // Rating out of 5
	Comment     string    `json:"comment" bson:"comment"` // Review comment
	Date        string    `json:"date" bson:"date"`       // Date of the review
	Likes       int       `json:"likes,omitempty" bson:"likes,omitempty"`
	Dislikes    int       `json:"dislikes,omitempty" bson:"dislikes,omitempty"`
	Attachments []Media   `json:"attachments,omitempty" bson:"attachments,omitempty"`
	CreatedAt   time.Time `json:"createdAt,omitempty" bson:"createdAt,omitempty"`
}

type Media struct {
	ID          string `json:"id" bson:"id"`
	EventID     string `json:"eventid" bson:"eventid"` // Reference to Event ID
	Type        string `json:"type" bson:"type"`       // e.g., "image", "video"
	URL         string `json:"url" bson:"url"`         // URL of the media
	Caption     string `json:"caption" bson:"caption"` // Optional caption for the media
	Description string `json:"description,omitempty"`
}

type Category struct {
	MainCategory  string   `json:"mainCategory,omitempty" bson:"mainCategory,omitempty"`
	SubCategories []string `json:"subCategories,omitempty" bson:"subCategories,omitempty"`
}

type Coordinates struct {
	Latitude  float64 `json:"latitude,omitempty" bson:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty" bson:"longitude,omitempty"`
}

type CheckIn struct {
	UserID    string    `json:"userId,omitempty" bson:"userId,omitempty"`
	PlaceID   string    `json:"placeId,omitempty" bson:"placeId,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
	Comment   string    `json:"comment,omitempty" bson:"comment,omitempty"`
	Rating    float64   `json:"rating,omitempty" bson:"rating,omitempty"` // Optional
	Medias    []Media   `json:"images,omitempty" bson:"images,omitempty"` // Optional
}

type PlaceVersion struct {
	PlaceID   string            `json:"placeId,omitempty" bson:"placeId,omitempty"`
	Version   int               `json:"version,omitempty" bson:"version,omitempty"`
	Data      Place             `json:"data,omitempty" bson:"data,omitempty"`
	UpdatedAt time.Time         `json:"updatedAt,omitempty" bson:"updatedAt,omitempty"`
	UpdatedBy string            `json:"updatedBy,omitempty" bson:"updatedBy,omitempty"`
	Changes   map[string]string `json:"changes,omitempty" bson:"changes,omitempty"`
}

type OperatingHours struct {
	Day          []string `json:"day,omitempty" bson:"day,omitempty"`
	OpeningHours []string `json:"opening,omitempty" bson:"opening,omitempty"`
	ClosingHours []string `json:"closing,omitempty" bson:"closing,omitempty"`
	TimeZone     string   `json:"timeZone,omitempty" bson:"timeZone,omitempty"`
}

type Tag struct {
	ID     string   `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string   `json:"name,omitempty" bson:"name,omitempty"`
	Places []string `json:"places,omitempty" bson:"places,omitempty"` // List of Place IDs tagged with this keyword
}

type Ticket struct {
	TicketID string  `json:"ticketid" bson:"ticketid"`
	EventID  string  `json:"eventid" bson:"eventid"`
	Name     string  `json:"name" bson:"name"`
	Price    float64 `json:"price" bson:"price"`
	Quantity int     `json:"quantity" bson:"quantity"`
}

const (
	PlaceStatusActive     = "active"
	PlaceStatusClosed     = "closed"
	PlaceStatusRenovation = "under renovation"
)

const (
	MediaTypeImage    = "image"
	MediaTypeVideo    = "video"
	MediaTypePhoto360 = "photo360"
)
