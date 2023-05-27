package main

import (
	"testing"
	"time"

	"github.com/mattermost/mattermost-server/server/v8/model"
	"github.com/stretchr/testify/assert"
)

var posts = []*model.Post{
	{
		Id:       "1234",
		UserId:   "1234",
		Message:  "@channel lala",
		CreateAt: time.Date(2023, 5, 23, 5, 0, 30, 0, time.UTC).UnixNano(),
	},
	{
		Id:       "1235",
		UserId:   "1234",
		Message:  "random message",
		CreateAt: time.Date(2023, 5, 23, 5, 30, 30, 0, time.UTC).UnixNano(),
	},
	{
		Id:       "1236",
		UserId:   "1234",
		Message:  "random message #2",
		CreateAt: time.Date(2023, 5, 23, 5, 15, 0, 0, time.UTC).UnixNano(),
	},
	{
		Id:       "1237",
		UserId:   "1234",
		Message:  "@here please ignore this message",
		CreateAt: time.Date(2023, 5, 23, 4, 45, 0, 40, time.UTC).UnixNano(),
	},
}

func TestSortPostsByCreationDate(t *testing.T) {
	tests := []struct {
		Name           string
		Payload        []*model.Post
		ExpectedOutput interface{}
	}{
		{
			Name:    "Sort by creation date not empty",
			Payload: posts,
			ExpectedOutput: []*model.Post{
				{
					Id:       "1237",
					Message:  "@here please ignore this message",
					CreateAt: time.Date(2023, 5, 23, 4, 45, 0, 40, time.UTC).UnixNano(),
				},
				{
					Id:       "1234",
					Message:  "@channel lala",
					CreateAt: time.Date(2023, 5, 23, 5, 0, 30, 0, time.UTC).UnixNano(),
				},
				{
					Id:       "1236",
					Message:  "random message #2",
					CreateAt: time.Date(2023, 5, 23, 5, 15, 0, 0, time.UTC).UnixNano(),
				},
				{
					Id:       "1235",
					Message:  "random message",
					CreateAt: time.Date(2023, 5, 23, 5, 30, 30, 0, time.UTC).UnixNano(),
				},
			},
		},
		{
			Name:           "Test Empty array",
			Payload:        []*model.Post{},
			ExpectedOutput: []*model.Post{},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			sortPostsByCreationDate(test.Payload)
			assert.ObjectsAreEqual(test.Payload, test.ExpectedOutput)
		})
	}
}

func TestFilterUserChannelMentions(t *testing.T) {
	tests := []struct {
		Name           string
		Payload        []*model.Post
		ExpectedOutput []*model.Post
	}{
		{
			Name:    "Expect two posts to contain channel wide mentions",
			Payload: posts,
			ExpectedOutput: []*model.Post{
				{
					Id:       "1237",
					UserId:   "1234",
					Message:  "@here please ignore this message",
					CreateAt: time.Date(2023, 5, 23, 4, 45, 0, 40, time.UTC).UnixNano(),
				},
				{
					Id:       "1234",
					UserId:   "1234",
					Message:  "@channel lala",
					CreateAt: time.Date(2023, 5, 23, 5, 0, 30, 0, time.UTC).UnixNano(),
				},
			},
		},
		{
			Name: "Given a different userId, Expect no posts to contain channel wide mentions",
			Payload: []*model.Post{
				{
					Id:       "1234",
					UserId:   "1235",
					Message:  "@channel lala",
					CreateAt: time.Date(2023, 5, 23, 4, 45, 0, 40, time.UTC).UnixNano(),
				},
			},
			ExpectedOutput: []*model.Post{},
		},
		{
			Name: "Given the same UserId and not channel wide mentions, expect no post to contain channel wide mentions",
			Payload: []*model.Post{
				{
					Id:       "1234",
					UserId:   "1234",
					Message:  "Doesn't contain channel wide mentions",
					CreateAt: time.Date(2023, 5, 23, 4, 45, 0, 40, time.UTC).UnixNano(),
				},
			},
			ExpectedOutput: []*model.Post{},
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			filtered := filterUserChannelMentions(test.Payload, "1234")
			assert.Equal(t, len(filtered), len(test.ExpectedOutput))
			assert.ObjectsAreEqual(filtered, test.ExpectedOutput)
		})
	}
}
