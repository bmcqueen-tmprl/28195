package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/converter"
	"go.temporal.io/sdk/workflow"
)

// main creates a new Schedule with the header {"TEST-HEADER": "TEST-VALUE"}
func main() {
	ppgtr := NewContextPropagator()
	c, err := client.Dial(client.Options{
		HostPort: "temporal.internal.k8s:7233",
		ConnectionOptions: client.ConnectionOptions{
			TLS: &tls.Config{},
		},
		ContextPropagators: []workflow.ContextPropagator{ppgtr},
	})
	if err != nil {
		log.Fatalln("Unable to create client", err)
	}
	defer c.Close()

	ctx := context.WithValue(context.Background(), PropagateKey, Values{Key: "TEST-HEADER", Value: "TEST-VALUE"})

	_, err = c.ScheduleClient().Create(ctx, client.ScheduleOptions{
		ID: "schedule2",
		Spec: client.ScheduleSpec{
			Calendars: []client.ScheduleCalendarSpec{
				{
					DayOfWeek: []client.ScheduleRange{
						{
							Start: 5,
							End:   5,
						},
					},
				},
			},
		},
		Action: &client.ScheduleWorkflowAction{
			Workflow:  "SomeWorkflow",
			TaskQueue: "SomeTaskQueue",
		},
	})
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

type (
	// contextKey is an unexported type used as key for items stored in the
	// Context object
	contextKey struct{}

	// propagator implements the custom context propagator
	propagator struct{}

	// Values is a struct holding values
	Values struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
)

// PropagateKey is the key used to store the value in the Context object
var PropagateKey = contextKey{}

// HeaderKey is the key used by the propagator to pass values through the
// Temporal server headers
const HeaderKey = "custom-header"

// NewContextPropagator returns a context propagator that propagates a set of
// string key-value pairs across a workflow
func NewContextPropagator() workflow.ContextPropagator {
	return &propagator{}
}

// Inject injects values from context into headers for propagation
func (s *propagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	value := ctx.Value(PropagateKey)
	payload, err := converter.GetDefaultDataConverter().ToPayload(value)
	if err != nil {
		return err
	}
	writer.Set(HeaderKey, payload)
	return nil
}

// InjectFromWorkflow injects values from context into headers for propagation
func (s *propagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	value := ctx.Value(PropagateKey)
	payload, err := converter.GetDefaultDataConverter().ToPayload(value)
	if err != nil {
		return err
	}
	writer.Set(HeaderKey, payload)
	return nil
}

// Extract extracts values from headers and puts them into context
func (s *propagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	if value, ok := reader.Get(HeaderKey); ok {
		var values Values
		if err := converter.GetDefaultDataConverter().FromPayload(value, &values); err != nil {
			return ctx, nil
		}
		ctx = context.WithValue(ctx, PropagateKey, values)
	}

	return ctx, nil
}

// Extract extracts values from headers and puts them into context
func (s *propagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	if value, ok := reader.Get(HeaderKey); ok {
		var values Values
		if err := converter.GetDefaultDataConverter().FromPayload(value, &values); err != nil {
			return ctx, nil
		}

		ctx = workflow.WithValue(ctx, PropagateKey, values)
	}

	return ctx, nil
}
