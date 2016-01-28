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

type Config struct {
	MsgDuration    time.Duration   `json:"msgDuration,omitempty"`
	IterationSleep time.Duration   `json:"iterationSleep,omitempty"`
	MaxDuration    time.Duration   `json:"maxDuration,omitempty"`
	MaxIterations  *int            `json:"maxIterations,omitempty"`
	BatchSize      int             `json:"batchSize,omitempty"`
	KeepAlive      bool            `json:"keepAlive,omitempty"`
	QueueName      string          `json:"queueName"`
	Env            config.Settings `json:"env"`
}

func DefaultConfig() *Config {
	return &Config{
		MsgDuration:    1 * time.Second,
		IterationSleep: 1 * time.Second,
		MaxDuration:    45 * time.Minute,
		BatchSize:      10,
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
	worker.ParseFlags()

	c := DefaultConfig()
	worker.ConfigFromJSON(c)

	if err := c.Valid(); err != nil {
		log.Fatal(err)
	}

	copy := *c
	copy.Env.Token = obfuscate(copy.Env.Token)
	j, err := json.MarshalIndent(copy, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(string(j))

	s := config.ManualConfig("iron_mq", &c.Env)
	t := time.Now()
	q := mq.ConfigNew(c.QueueName, &s)
	info, err := q.Info()
	if err != nil {
		log.Fatal("Could not access queue info", err)
	}
	fmt.Printf("Queue has %d messages. (request took %v)\n", info.Size, time.Since(t))

batchLoop:
	for x := 0; c.MaxIterations == nil || x < *c.MaxIterations; x++ {
		// Determine if we have time for another batch
		if time.Since(start) > c.MaxDuration {
			break
		}

		if x != 0 {
			fmt.Println("Sleeping", c.IterationSleep)
			time.Sleep(c.IterationSleep)
		}

		// Create a timeout that can accommodate the duration of the entire batch
		timeout := float64(c.BatchSize) * c.MsgDuration.Seconds() * 1.1

		// Reserve messages
		t = time.Now()
		msgs, err := q.LongPoll(c.BatchSize, int(timeout), 0, false)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Iteration %d: Requested %d, got %d (request took %v)\n", x, c.BatchSize, len(msgs), time.Since(t))

		if len(msgs) == 0 && !c.KeepAlive {
			fmt.Println("Queue is empty - breaking work loop")
			break
		}

		for i, msg := range msgs {
			// Determine if we have time for another message
			if time.Since(start) > c.MaxDuration+c.MsgDuration {
				fmt.Println("Not enough time to process message - breaking work loop")
				break batchLoop
			}

			fmt.Printf(" %d: %q\n", i, msg.Body)

			// Simulate Message Processing
			time.Sleep(c.MsgDuration)

			err = msg.Delete()
			if err != nil {
				fmt.Println("Could not delete msg:", msg.Body, err)
			}
		}
	}

	fmt.Println("Worker ending after", time.Since(start))
}

func obfuscate(v string) string {
	r := []rune(v)
	for x := len(r) / 4; x < len(r); x++ {
		r[x] = '*'
	}
	return string(r)
}
