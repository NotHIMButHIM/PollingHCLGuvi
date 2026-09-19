package models

import "go.mongodb.org/mongo-driver/bson/primitive"

type Option struct {
	ID   string `bson:"id" json:"id"`
	Text string `bson:"text" json:"text"`
}

type Poll struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Question  string             `bson:"question" json:"question"`
	Options   []Option           `bson:"options" json:"options"`
	CreatedBy primitive.ObjectID `bson:"createdBy" json:"createdBy"`
}

type Vote struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	PollID   string             `bson:"pollId" json:"pollId"`
	OptionID string             `bson:"optionId" json:"optionId"`
}
