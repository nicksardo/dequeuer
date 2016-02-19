# dequeuer
Configurable Test Worker for Iron.io

### Assumptions
- You have the Iron.io CLI installed and have a `iron.json` in the local directory or `.iron.json` in your home directory.
- You have docker installed

### Build
Cross compile or use docker to generate linux/amd64 executable
```shell
❯❯❯ docker run --rm -it -v $PWD:/go/src/a -w /go/src/a iron/go:dev sh -c 'go get ./... && go build -o dequeuer main.go'
```

##### Option A: Deploy by bundling into docker image
###### 1. Build and push docker image
```shell
docker build -t {{youraccount}}/dequeuer:0.1 .
docker push {{youraccount}}/dequeuer:0.1
```

###### 2. Test again by creating a config file and running your docker image
```shell
vi prod_config.json  # create a config file called "prod_config.json" in an empty directory
docker run --rm -it -e "CONFIG_FILE=prod_config.json" -v $PWD:/app {{youraccount}}/dequeuer
```

###### 3.  Register your docker image with Iron.io
```shell
iron register {{youraccount}}/dequeuer:0.1
```

##### Option B: Deploy by creating and uploading zip file
###### 1. Zip and upload to Iron
```shell
❯❯❯ zip -r dequeuer.zip dequeuer
❯❯❯ iron worker upload -zip dequeuer.zip -name dequeuer iron/base ./dequeuer  # Uses `iron/base` as docker image.  
----->  Configuring client
        Project 'Spinnaker 1' with id='XXXXX'
----->  Uploading worker 'dequeuer'
        Uploaded code package with id='ABC'
        Check https://hud.iron.io/tq/projects/XXX/code/YYY for more info
```

#### Set configuration data
Visit [hud.iron.io](https://hud.iron.io) and modify the configuration for the `nicksardo/dequeuer` code package.

### IronWorker Config
```javascript
{
  "msgDuration": 1000000000,     // optional, default shown, nanoseconds, time to sleep per message
  "iterationSleep": 20000000,    // optional, default shown, nanoseconds, time to sleep between batches
  "maxDuration": 2700000000000,  // optional, default shown, nanoseconds, max time for worker to live
  "batchSize": 1,                // optional, default shown, size of batch
  "maxEmptyResults": 0,          // optional, default shown, kills worker after X sequential empty queue results
  "maxIterations": null,         // optional, default shown, max number of batch requests
  "dequeueWait": 0,              // optional, default shown, long poll wait (0-30 seconds)
  "queueName": "sampleQueue",    // required
  "env": {                       
    "project_id": "XXX",         // required
    "host": "abc.iron.io",       // required
    "token": "YYY"               // required
  }
}
```

### Execution
```shell
# Create messages onto the queue
❯❯❯ iron mq push sampleQueue a b c d e f g h i j k l m n o p q r s t u v w x y z
-----> Configuring client
       Project 'Spinnaker 1' with id='ABC'
-----> Message succesfully pushed!
       Message IDs:
       6244731235502737656 6244731235502737657 6244731235502737658 6244731235502737659 6244731235502737660 6244731235502737661 6244731235502737662 6244731235502737663 6244731235502737664 6244731235502737665 6244731235502737666 6244731235502737667 6244731235502737668 6244731235502737669 6244731235502737670 6244731235502737671 6244731235502737672 6244731235502737673 6244731235502737674 6244731235502737675 6244731235502737676 6244731235502737677 6244731235502737678 6244731235502737679 6244731235502737680 6244731235502737681


# Process the messages
❯❯❯ iron worker queue --wait nicksardo/dequeuer
----->  Configuring client
        Project 'Spinnaker 1' with id='XXX'
----->  Queueing task 'nicksardo/dequeuer'
        Queued task with id='ABC'
        Check https://hud.iron.io/tq/projects/XXX/jobs/YYY for more info
----->  Waiting for task56a9be008ba9d6000601524a
----->  Done
----->  Printing Log:
{
    "msgDuration": 1000000000,
    "iterationSleep": 1000000000,
    "maxDuration": 2700000000000,
    "batchSize": 10,
    "queueName": "sampleQueue",
    "env": {
        "token": "YYY",
        "project_id": "XXX",
        "host": "mq-aws-us-east-1-1.iron.io"
    }
}
Queue has 26 messages.
Iteration 0: Requested 10, got 10
 0: "a"
 1: "b"
 2: "c"
 3: "d"
 4: "e"
 5: "f"
 6: "g"
 7: "h"
 8: "i"
 9: "j"
Sleeping 1s
Iteration 1: Requested 10, got 10
 0: "k"
 1: "l"
 2: "m"
 3: "n"
 4: "o"
 5: "p"
 6: "q"
 7: "r"
 8: "s"
 9: "t"
Sleeping 1s
Iteration 2: Requested 10, got 6
 0: "u"
 1: "v"
 2: "w"
 3: "x"
 4: "y"
 5: "z"
Sleeping 1s
Iteration 3: Requested 10, got 0
Queue is empty - breaking work loop
Worker ending
```
