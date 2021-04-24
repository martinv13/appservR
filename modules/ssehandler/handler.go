// Adapted from Gin SSE example

package ssehandler

import (
	"io"

	"github.com/gin-gonic/gin"
)

//It keeps a list of clients those are currently attached
//and broadcasting events to those clients.
type Event struct {

	// Events are pushed to this channel by the main events-gathering routine
	Message chan string

	// Keep the last message broadcasted to send it to new clients
	LastMessage string

	// New client connections
	NewClients chan chan string

	// Closed client connections
	ClosedClients chan chan string

	// Total client connections
	TotalClients map[chan string]bool
}

// Initialize event and Start procnteessing requests
func NewServer() (event *Event) {

	event = &Event{
		Message:       make(chan string),
		NewClients:    make(chan chan string),
		ClosedClients: make(chan chan string),
		TotalClients:  make(map[chan string]bool),
	}

	go event.listen()

	return
}

// Get controller
func (stream *Event) Controller() gin.HandlerFunc {

	return func(c *gin.Context) {

		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")

		clientChan := make(chan string)

		// Send new connection to event server
		stream.NewClients <- clientChan

		defer func() {
			// Send closed connection to event server
			stream.ClosedClients <- clientChan
		}()

		go func() {
			// Send connection that is closed by client to event server
			<-c.Done()
			stream.ClosedClients <- clientChan
		}()

		c.Stream(func(w io.Writer) bool {
			// Stream message to client from message channel
			if msg, ok := <-clientChan; ok {
				c.SSEvent("message", msg)
				return true
			}
			return false
		})
	}

}

//Handles addition and removal of clients and broadcast messages to clients.
func (stream *Event) listen() {
	for {
		select {
		// Add new available client and send last message
		case client := <-stream.NewClients:
			stream.TotalClients[client] = true
			client <- stream.LastMessage

		// Remove closed client
		case client := <-stream.ClosedClients:
			delete(stream.TotalClients, client)

		// Save last message and broadcast it to client
		case eventMsg := <-stream.Message:
			if stream.LastMessage != eventMsg {
				stream.LastMessage = eventMsg
				for clientMessageChan := range stream.TotalClients {
					clientMessageChan <- eventMsg
				}
			}
		}
	}
}
