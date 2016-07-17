## Minimalistic Stressing Framework

The `stress` package aims to provide a way to simulate work in a stressed environment.

### Basic Usage

```go
stresser := &stress.Stresser{
	New: func() stress.Worker {
		return &MyStressWorker{}
	},
	Limit: rate.NewLimiter(eventsPerSecond, eventsPerSecond)
}

go stresser.Start()

type MyStressWorker struct {}

func (worker *MyStressWorker) Work(ctx context.Context) {
	// Perform a single unit of work.
}

```

### Examples

 - `queue-producer` - Pumps messages into an AMQP exchange.

### TODO

 - Expand testing
 - Add more examples

