package controller

import (
	"Tools3-Project/models"
	"context"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var eventCollection *mongo.Collection

func InitEventCollection(db *mongo.Database) {
	eventCollection = db.Collection("events")
}

func eventToMap(ev models.Event) gin.H {
	return gin.H{
		"id":          ev.ID.Hex(),
		"title":       ev.Title,
		"date":        ev.Date,
		"time":        ev.Time,
		"location":    ev.Location,
		"description": ev.Description,
		"organizer":   ev.Organizer,
		"attendees":   ev.Attendees,
	}
}

func CreateEvent(c *gin.Context) {
	session := sessions.Default(c)
	email := session.Get("user")
	if email == nil {
		c.JSON(401, gin.H{"error": "Not logged in"})
		return
	}

	var event models.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	event.ID = primitive.NewObjectID()
	event.Organizer = email.(string)
	event.Attendees = []models.Attendee{
		{Email: email.(string), Status: "organizer"},
	}

	_, err := eventCollection.InsertOne(context.Background(), event)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to create event"})
		return
	}

	c.JSON(200, gin.H{"message": "Event created", "eventId": event.ID.Hex()})
}

func GetMyOrganizedEvents(c *gin.Context) {
	session := sessions.Default(c)
	email := session.Get("user")
	if email == nil {
		c.JSON(401, gin.H{"error": "Not logged in"})
		return
	}

	cursor, err := eventCollection.Find(context.Background(), bson.M{"organizer": email.(string)})
	if err != nil {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}
	var events []models.Event
	if err := cursor.All(context.Background(), &events); err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse events"})
		return
	}

	out := make([]gin.H, 0, len(events))
	for _, ev := range events {
		out = append(out, eventToMap(ev))
	}
	c.JSON(200, out)
}

func GetMyInvitedEvents(c *gin.Context) {
	session := sessions.Default(c)
	email := session.Get("user")
	if email == nil {
		c.JSON(401, gin.H{"error": "Not logged in"})
		return
	}

	filter := bson.M{
		"attendees": bson.M{
			"$elemMatch": bson.M{
				"email":  email.(string),
				"status": "pending",
			},
		},
		"organizer": bson.M{"$ne": email.(string)},
	}

	cursor, err := eventCollection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}

	var events []models.Event
	if err := cursor.All(context.Background(), &events); err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse events"})
		return
	}

	out := make([]gin.H, 0, len(events))
	for _, ev := range events {
		out = append(out, eventToMap(ev))
	}

	c.JSON(200, out)
}

func GetMyAcceptedEvents(c *gin.Context) {
	session := sessions.Default(c)
	email := session.Get("user")
	if email == nil {
		c.JSON(401, gin.H{"error": "Not logged in"})
		return
	}

	filter := bson.M{
		"attendees": bson.M{
			"$elemMatch": bson.M{
				"email":  email.(string),
				"status": "accepted",
			},
		},
		"organizer": bson.M{"$ne": email.(string)},
	}

	cursor, err := eventCollection.Find(context.Background(), filter)
	if err != nil {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}

	var events []models.Event
	if err := cursor.All(context.Background(), &events); err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse events"})
		return
	}

	out := make([]gin.H, 0, len(events))
	for _, ev := range events {
		out = append(out, eventToMap(ev))
	}
	c.JSON(200, out)
}

func InviteUser(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get("user")
	if user == nil {
		c.JSON(401, gin.H{"error": "Not logged in"})
		return
	}

	var input models.InviteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "Invalid input"})
		return
	}

	objID, err := primitive.ObjectIDFromHex(input.EventID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid event ID"})
		return
	}

	_, err = eventCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID},
		bson.M{"$addToSet": bson.M{"attendees": models.Attendee{Email: input.Email, Status: "pending"}}},
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to invite user"})
		return
	}

	var ev models.Event
	if err := eventCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&ev); err != nil {
		c.JSON(500, gin.H{"error": "Invite applied but failed to fetch event"})
		return
	}

	c.JSON(200, gin.H{"message": "User invited", "event": eventToMap(ev)})
}

func RespondToEvent(c *gin.Context) {
	var input models.InviteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	session := sessions.Default(c)
	user := session.Get("user")
	if user == nil {
		c.JSON(401, gin.H{"error": "Not logged in"})
		return
	}
	userEmail := user.(string)

	objID, err := primitive.ObjectIDFromHex(input.EventID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid event ID"})
		return
	}

	var newStatus string
	switch input.Status {
	case "going", "accepted", "accept":
		newStatus = "accepted"
	case "not_going", "declined", "decline":
		newStatus = "declined"
	case "pending", "":
		newStatus = "pending"
	default:
		c.JSON(400, gin.H{"error": "Invalid status"})
		return
	}

	res, err := eventCollection.UpdateOne(
		context.Background(),
		bson.M{"_id": objID, "attendees.email": userEmail},
		bson.M{"$set": bson.M{"attendees.$.status": newStatus}},
	)
	if err != nil {
		c.JSON(500, gin.H{"error": "Failed to update response"})
		return
	}

	if res.MatchedCount == 0 {
		c.JSON(404, gin.H{"error": "You are not an attendee of this event"})
		return
	}

	var ev models.Event
	if err := eventCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&ev); err != nil {
		c.JSON(200, gin.H{"message": "Response recorded", "status": newStatus})
		return
	}

	c.JSON(200, gin.H{"message": "Response recorded", "status": newStatus, "event": eventToMap(ev)})
}

func GetEventAttendees(c *gin.Context) {
	eventID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid event ID"})
		return
	}

	session := sessions.Default(c)
	user := session.Get("user")
	if user == nil {
		c.JSON(401, gin.H{"error": "Not logged in"})
		return
	}

	var event models.Event
	if err := eventCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&event); err != nil {
		c.JSON(404, gin.H{"error": "Event not found"})
		return
	}

	out := make([]gin.H, 0, len(event.Attendees))
	for _, a := range event.Attendees {
		if a.Status == "accepted" || a.Status == "going" {
			local := a.Email
			if parts := strings.Split(a.Email, "@"); len(parts) > 0 {
				local = parts[0]
			}
			name := strings.ReplaceAll(local, ".", " ")
			name = strings.ReplaceAll(name, "_", " ")
			words := strings.Fields(name)
			for i, w := range words {
				if len(w) > 0 {
					words[i] = strings.ToUpper(string(w[0])) + strings.ToLower(w[1:])
				}
			}
			displayName := strings.Join(words, " ")

			out = append(out, gin.H{
				"inviteId": a.Email,
				"name":     displayName,
				"email":    a.Email,
				"status":   a.Status,
			})
		}
	}

	c.JSON(200, out)
}

func DeleteEvent(c *gin.Context) {
	session := sessions.Default(c)
	email := session.Get("user")
	if email == nil {
		c.JSON(401, gin.H{"error": "Not logged in"})
		return
	}

	id := c.Param("id")
	objID, _ := primitive.ObjectIDFromHex(id)

	res, err := eventCollection.DeleteOne(context.Background(),
		bson.M{"_id": objID, "organizer": email.(string)})

	if err != nil {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}

	if res.DeletedCount == 0 {
		c.JSON(403, gin.H{"error": "Not allowed"})
		return
	}

	c.JSON(200, gin.H{"message": "Event deleted"})
}

func SearchEvents(c *gin.Context) {
	q := c.Query("q")
	date := c.Query("date")
	session := sessions.Default(c)
	user := session.Get("user")

	searchFilter := bson.M{}

	if q != "" {
		searchFilter["$or"] = []bson.M{
			{"title": bson.M{"$regex": q, "$options": "i"}},
			{"description": bson.M{"$regex": q, "$options": "i"}},
		}
	}

	if date != "" {
		searchFilter["date"] = date
	}

	if user != nil {
		searchFilter = bson.M{
			"$and": []bson.M{
				searchFilter,
				{"attendees": bson.M{"$not": bson.M{"$elemMatch": bson.M{"email": user.(string), "status": "declined"}}}},
			},
		}
	}

	cursor, err := eventCollection.Find(context.Background(), searchFilter)
	if err != nil {
		c.JSON(500, gin.H{"error": "Database error"})
		return
	}

	var events []models.Event
	if err := cursor.All(context.Background(), &events); err != nil {
		c.JSON(500, gin.H{"error": "Failed to parse events"})
		return
	}

	out := make([]gin.H, 0, len(events))
	for _, ev := range events {
		out = append(out, eventToMap(ev))
	}

	c.JSON(200, out)
}

func GetEventByID(c *gin.Context) {
	eventID := c.Param("id")
	objID, err := primitive.ObjectIDFromHex(eventID)
	if err != nil {
		c.JSON(400, gin.H{"error": "Invalid event ID"})
		return
	}

	session := sessions.Default(c)
	user := session.Get("user")
	if user == nil {
		c.JSON(401, gin.H{"error": "Not logged in"})
		return
	}

	var ev models.Event
	if err := eventCollection.FindOne(context.Background(), bson.M{"_id": objID}).Decode(&ev); err != nil {
		c.JSON(404, gin.H{"error": "Event not found"})
		return
	}

	c.JSON(200, eventToMap(ev))
}
