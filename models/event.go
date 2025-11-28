package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Attendee struct {
    Email  string `bson:"email" json:"email"`
    Status string `bson:"status" json:"status"`
}

type Event struct {
    ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
    Title       string             `bson:"title" json:"title"`
    Date        string             `bson:"date" json:"date"`
    Time        string             `bson:"time" json:"time"`
    Location    string             `bson:"location" json:"location"`
    Description string             `bson:"description" json:"description"`

    Organizer   string     `bson:"organizer" json:"organizer"`
    Attendees   []Attendee `bson:"attendees" json:"attendees"`
}

