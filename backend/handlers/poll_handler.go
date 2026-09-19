package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"pollingapp/backend/config"
	"pollingapp/backend/models"
)

type createPollInput struct {
	Question string   `json:"question" binding:"required"`
	Options  []string `json:"options" binding:"required"`
}

type voteInput struct {
	OptionID string `json:"optionId" binding:"required"`
}

func CreatePoll(c *gin.Context) {
	var input createPollInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(input.Options) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least 2 options required"})
		return
	}

	userIdStr := c.GetString("userId")
	userId, err := primitive.ObjectIDFromHex(userIdStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	var options []models.Option
	for i, text := range input.Options {
		if text == "" {
			continue
		}
		options = append(options, models.Option{
			ID:   strconv.Itoa(i + 1),
			Text: text,
		})
	}

	if len(options) < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "at least 2 non empty options required"})
		return
	}

	poll := models.Poll{
		Question:  input.Question,
		Options:   options,
		CreatedBy: userId,
	}

	polls := config.DB.Collection("polls")
	res, err := polls.InsertOne(config.Ctx, poll)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not create poll"})
		return
	}

	pollId := res.InsertedID.(primitive.ObjectID).Hex()

	for _, opt := range options {
		config.RedisClient.HSet(config.Ctx, "poll:"+pollId+":votes", opt.ID, 0)
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       pollId,
		"question": poll.Question,
		"options":  options,
	})
}

func GetPoll(c *gin.Context) {
	id := c.Param("id")

	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
		return
	}

	polls := config.DB.Collection("polls")

	var poll models.Poll
	err = polls.FindOne(config.Ctx, bson.M{"_id": objId}).Decode(&poll)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
		return
	}

	counts, err := config.RedisClient.HGetAll(config.Ctx, "poll:"+id+":votes").Result()
	if err != nil {
		counts = map[string]string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"id":       id,
		"question": poll.Question,
		"options":  poll.Options,
		"counts":   counts,
	})
}

func Vote(c *gin.Context) {
	id := c.Param("id")

	objId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid poll id"})
		return
	}

	var input voteInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	polls := config.DB.Collection("polls")

	var poll models.Poll
	err = polls.FindOne(config.Ctx, bson.M{"_id": objId}).Decode(&poll)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "poll not found"})
		return
	}

	valid := false
	for _, opt := range poll.Options {
		if opt.ID == input.OptionID {
			valid = true
			break
		}
	}

	if !valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid option"})
		return
	}

	votesKey := "poll:" + id + ":votes"

	_, err = config.RedisClient.HIncrBy(config.Ctx, votesKey, input.OptionID, 1).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not register vote"})
		return
	}

	votes := config.DB.Collection("votes")
	votes.InsertOne(config.Ctx, models.Vote{
		PollID:   id,
		OptionID: input.OptionID,
	})

	allCounts, err := config.RedisClient.HGetAll(config.Ctx, votesKey).Result()
	if err != nil {
		allCounts = map[string]string{}
	}

	PublishUpdate(id, gin.H{
		"counts": allCounts,
	})

	c.JSON(http.StatusOK, gin.H{"message": "vote recorded"})
}
