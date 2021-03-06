package tracer

import (
	"time"

	"github.com/epsagon/epsagon-go/protocol"
)

// MockedEpsagonTracer will not send traces if closed
type MockedEpsagonTracer struct {
	Exceptions      *[]*protocol.Exception
	Events          *[]*protocol.Event
	Labels          map[string]interface{}
	RunnerException *protocol.Exception
	Config          *Config

	PanicStart        bool
	PanicAddEvent     bool
	PanicAddException bool
	PanicStop         bool
	DelayAddEvent     bool
	DelayedEventsChan chan bool
	stopped           bool
}

// Start implementes mocked Start
func (t *MockedEpsagonTracer) Start() {
	if t.PanicStart {
		panic("panic in Start()")
	}
	t.stopped = false
}

// Running implementes mocked Running
func (t *MockedEpsagonTracer) Running() bool {
	return false
}

// Stop implementes mocked Stop
func (t *MockedEpsagonTracer) Stop() {
	if t.PanicStop {
		panic("panic in Stop()")
	}
	t.stopped = true
}

// Stopped implementes mocked Stopped
func (t *MockedEpsagonTracer) Stopped() bool {
	return t.stopped
}

// AddEvent implementes mocked AddEvent
func (t *MockedEpsagonTracer) AddEvent(e *protocol.Event) {
	if t.PanicAddEvent {
		panic("panic in AddEvent()")
	}
	if t.DelayAddEvent {
		go func() {
			time.Sleep(time.Second)
			*t.Events = append(*t.Events, e)
			t.DelayedEventsChan <- true
		}()
	} else {
		*t.Events = append(*t.Events, e)
	}
}

// AddException implementes mocked AddEvent
func (t *MockedEpsagonTracer) AddException(e *protocol.Exception) {
	if t.PanicAddException {
		panic("panic in AddException()")
	}
	*t.Exceptions = append(*t.Exceptions, e)
}

// GetConfig implementes mocked AddEvent
func (t *MockedEpsagonTracer) GetConfig() *Config {
	return t.Config
}

// AddExceptionTypeAndMessage implements AddExceptionTypeAndMessage
func (t *MockedEpsagonTracer) AddExceptionTypeAndMessage(exceptionType, msg string) {
	t.AddException(&protocol.Exception{
		Type:    exceptionType,
		Message: msg,
		Time:    GetTimestamp(),
	})
}

// AddLabel implements AddLabel
func (t *MockedEpsagonTracer) AddLabel(key string, value interface{}) {
	t.Labels[key] = value
}

// verifyLabel implements verifyLabel
func (t *MockedEpsagonTracer) verifyLabel(label epsagonLabel) bool {
	return true
}

// AddError implements AddError
func (t *MockedEpsagonTracer) AddError(errorType string, value interface{}) {
	t.RunnerException = &protocol.Exception{
		Type:    errorType,
		Message: "test",
	}
}

// GetRunnerEvent implements AddError
func (t *MockedEpsagonTracer) GetRunnerEvent() *protocol.Event {
	for _, event := range *t.Events {
		if event.Origin == "runner" {
			return event
		}
	}
	return nil
}
