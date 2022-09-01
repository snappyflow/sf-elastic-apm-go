# sf-elastic-apm-go

This provides the instrumentation modules to trace your Go applications.


## Pre-requisite

Install the Elastic APM Go Agent with the following command.
```bash
go get go.elastic.co/apm/v2
```

## Instrumentation modules

**module/apmgoji**

Package apmgoji provides middleware for the Goji web framework (https://github.com/zenazn/goji). This middleware traces all the incoming requests and reports each transaction to the APM server.

The apmgoji middleware will also recover panics and send them to Elastic APM.

```go
import (
	"github.com/snappyflow/sf-elastic-apm-go/module/apmgoji"
)

func main() {
	goji.Use(goji.DefaultMux.Router)
	goji.Use(apmgoji.Middleware())
	...
}
```


**module/apmgoredisv8**

Package apmgoredisv8 provides a means of instrumenting go-redis/redis for v8 so that Redis commands are reported as spans within the current transaction.

```go
import (
	"github.com/go-redis/redis/v8"
	"github.com/snappyflow/sf-elastic-apm-go/module/apmgoredisv8"
)

var redisClient = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

func main() {
	// Add apm hook to redisClient.
	redisClient.AddHook(apmgoredis.NewHook())
	...
}

func handleRequest(w http.ResponseWriter, req *http.Request) {
	// Pass the current transaction context in Redis call
	// Redis commands will be reported as spans within the current transaction.
	val, err := redisClient.Get(req.Context(), "key1").Result()
	if err != nil {
		fmt.Println(err)
	}
	...
}
```


**module/apmmongo**

Package apmmongo provides a means of instrumenting the MongoDB Go Driver (https://github.com/mongodb/mongo-go-driver), so that MongoDB commands are reported as spans within the current transaction.

```go
import (
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.elastic.co/apm/module/apmmongo/v2"
)

var mongoClient, _ = mongo.Connect(
	context.Background(),
	options.Client().SetMonitor(apmmongo.CommandMonitor()).ApplyURI("mongodb://localhost:27017"),
)

func handleRequest(w http.ResponseWriter, req *http.Request) {
	// Pass the current transaction context in Redis call
	// Redis commands will be reported as spans within the current transaction.
	collection := mongoClient.Database("db").Collection("coll")
	cur, err := collection.Find(r.Context(), bson.D{})
	...
}
```