package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/iron-io/iron_go3/config"
	"github.com/iron-io/iron_go3/mq"
	"github.com/iron-io/iron_go3/worker"
)

// Configuration set via HUD
type Config struct {
	MsgDuration     time.Duration   `json:"msgDuration,omitempty"`
	IterationSleep  time.Duration   `json:"iterationSleep,omitempty"`
	MaxDuration     time.Duration   `json:"maxDuration,omitempty"`
	MaxIterations   *int            `json:"maxIterations,omitempty"`
	BatchSize       int             `json:"batchSize,omitempty"`
	MaxEmptyResults *int            `json:"maxEmptyResults,omitempty"`
	DequeueWait     int             `json:"dequeueWait,omitempty"`
	QueueName       string          `json:"queueName"`
	Env             config.Settings `json:"env"`
}

func DefaultConfig() *Config {
	m := 0
	return &Config{
		MsgDuration:     1 * time.Second,
		IterationSleep:  20 * time.Millisecond,
		MaxDuration:     45 * time.Minute,
		BatchSize:       1,
		DequeueWait:     0,
		MaxEmptyResults: &m,
	}
}

func (c *Config) Valid() error {
	if len(c.Env.ProjectId) == 0 || len(c.Env.Token) == 0 {
		return errors.New(`Require queue project & token set in code config.`)
	}
	return nil
}

func main() {
	start := time.Now()

	// Parse config json and validate
	worker.ParseFlags()
	c := DefaultConfig()
	worker.ConfigFromJSON(c)
	if err := c.Valid(); err != nil {
		log.Fatal(err)
	}

	// Copy config, obfuscate token, and print for record
	copy := *c
	copy.Env.Token = obfuscate(copy.Env.Token)
	j, err := json.MarshalIndent(copy, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(j))

	// Retrieve queue info
	s := config.ManualConfig("iron_mq", &c.Env)
	t := time.Now()
	q := mq.ConfigNew(c.QueueName, &s)
	info, err := q.Info()
	if err != nil {
		log.Fatal("Could not access queue info", err)
	}
	fmt.Printf("Queue has %d messages. (request took %v)\n", info.Size, time.Since(t))

	// Loop for multiple iterations of dequeuing & processing messages
	emptyResultCount := 0
	totalProcessedCount := 0
batchLoop:
	for x := 0; c.MaxIterations == nil || x < *c.MaxIterations; x++ {
		// Determine if we have time for another set of processing
		if time.Since(start) > c.MaxDuration {
			break
		}

		// If configured, sleep between iterations
		if x != 0 {
			fmt.Println("Sleeping", c.IterationSleep)
			time.Sleep(c.IterationSleep)
		}

		// Create a timeout that can accommodate the duration of the entire batch
		timeout := float64(c.BatchSize) * c.MsgDuration.Seconds() * 1.1 // give it an extra 10% time for safety

		// Reserve messages with given batch size, timeout, and longpoll time
		t = time.Now()
		msgs, err := q.LongPoll(c.BatchSize, int(timeout), c.DequeueWait, false)
		if err != nil {
			// Ideally, continue the loop and try again later. After a certain amount of time, do panic/Fatal
			log.Fatal(err)
		}
		fmt.Printf("Iteration %d: Requested %d, got %d (request took %v with a max wait of %ds)\n", x, c.BatchSize, len(msgs), time.Since(t), c.DequeueWait)

		// Handle case of zero messages
		if len(msgs) == 0 && c.MaxEmptyResults != nil {
			emptyResultCount++
			if emptyResultCount > *c.MaxEmptyResults {
				fmt.Println("Queue is empty - breaking work loop")
				break
			}
		} else {
			// Reset count if queue isn't empty
			emptyResultCount = 0
		}

		// Process each message
		for i, msg := range msgs {
			// Determine if we have time to process another message
			if time.Since(start) > c.MaxDuration+c.MsgDuration {
				fmt.Println("Not enough time to process message - breaking work loop")
				break batchLoop
			}

			// Simulate Message Processing
			time.Sleep(c.MsgDuration)
			fmt.Printf(" %d: %q\n", i, msg.Body)

			// Example case for error processing
			var processingError error
			if processingError != nil {
				// Move error to error queue
				// errorQueue.Push(msg)
			}

			totalProcessedCount++

			// Delete processed message from queue
			err = msg.Delete()
			if err != nil {
				fmt.Println("Could not delete msg:", msg.Body, err)
			}
		}
	}

	fmt.Printf("Worker ending after %s and processing %d messages", time.Since(start), totalProcessedCount)
}

func obfuscate(v string) string {
	r := []rune(v)
	for x := len(r) / 4; x < len(r); x++ {
		r[x] = '*'
	}
	return string(r)
}
