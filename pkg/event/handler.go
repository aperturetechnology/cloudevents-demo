/*
Copyright 2018 Google, Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package event

import (
	"fmt"
	"io"
	"net/http"
	"reflect"

	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
)

const (
	usage = "event.NewHandler expected a `func(<data>, event.Context) error`"

	googleLoadBalancerAgent = "GoogleHC/1.0"
)

type handler struct {
	fnValue  reflect.Value
	dataType reflect.Type
}

func assertEventHandler(fn interface{}) (dataType reflect.Type) {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		panic(usage + "; did not receive a func")
	}
	if fnType.NumIn() != 2 {
		panic(usage + "; wrong parameter count")
	}
	if !fnType.In(1).ConvertibleTo(reflect.TypeOf(&Context{})) {
		panic(usage + "; cannot convert " + fnType.In(1).Name() + " to event.Context")
	}
	if fnType.NumOut() != 1 {
		panic(usage + "; wrong output count")
	}
	// We have to use an awkward jump into and out of a pointer to avoid passing a literal
	// nil to reflect, which would lose all type information and assert.
	errorType := reflect.TypeOf((*error)(nil)).Elem()
	if !fnType.Out(0).ConvertibleTo(errorType) {
		panic(usage + "; cannot convert " + fnType.Out(0).Name() + " to error")
	}

	return fnType.In(0)
}

// Handler creates an EventHandler that implements http.Handler
// Will panic in case of a type error
// @param fn a function of type func(<your data struct>, *event.Context) error
// TODO(inlined): for continuations we'll probably change the return signature to (interface{}, error)
func Handler(fn interface{}) http.Handler {
	return &handler{dataType: assertEventHandler(fn), fnValue: reflect.ValueOf(fn)}
}

func (h *handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Respond OK to GKE loadbalancers
	if r.Header.Get("User-Agent") == googleLoadBalancerAgent {
		return
	}

	elemType := h.dataType
	if h.dataType.Kind() == reflect.Ptr {
		elemType = h.dataType.Elem()
	}
	dataPtrVal := reflect.New(elemType)

	context, err := FromRequest(dataPtrVal.Interface(), r)
	if err != nil {
		glog.Warning("Failed to handle request", spew.Sdump(r))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`Invalid request`))
		return
	}

	args := []reflect.Value{dataPtrVal, reflect.ValueOf(context)}
	if h.dataType.Kind() != reflect.Ptr {
		args[0] = args[0].Elem()
	}
	res := h.fnValue.Call(args)[0]
	if res.IsNil() {
		return
	}
	// Type cast safe due to assertEventHandler()
	err = res.Interface().(error)
	glog.Error("Failed to handle event: ", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`Internal server error`))
}

// Mux allows developers to handle logically related groups of
// functionality multiplexed based on the event type.
// BUG: Mux relies on JSON encoding for events.
type Mux map[string]*handler

// NewMux creates a new Mux
func NewMux() Mux {
	return make(map[string]*handler)
}

// Handle adds a new handler for a specific event type
func (m Mux) Handle(eventType string, fn interface{}) {
	m[eventType] = &handler{dataType: assertEventHandler(fn), fnValue: reflect.ValueOf(fn)}
}

// ServeHTTP implements http.Handler
func (m Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Respond OK to GKE loadbalancers
	if r.Header.Get("User-Agent") == googleLoadBalancerAgent {
		return
	}

	var rawData io.Reader
	context, err := FromRequest(&rawData, r)
	if err != nil {
		glog.Warning("Failed to handle request", spew.Sdump(r))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`Invalid request`))
		return
	}

	h := m[context.EventType]
	if h == nil {
		glog.Warning("Cloud not find handler for event type", context.EventType)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Event type %q is not supported", context.EventType)))
		return
	}

	elemType := h.dataType
	if h.dataType.Kind() == reflect.Ptr {
		elemType = h.dataType.Elem()
	}
	dataPtrVal := reflect.New(elemType)
	parseRes := reflect.ValueOf(unmarshalEventData).Call([]reflect.Value{
		reflect.ValueOf(context.ContentType),
		reflect.ValueOf(rawData),
		dataPtrVal,
	})[0]
	if !parseRes.IsNil() {
		err := parseRes.Interface().(error)
		glog.Warning("Failed to parse event data", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`Invalid request`))
		return
	}

	args := []reflect.Value{dataPtrVal, reflect.ValueOf(context)}
	if h.dataType.Kind() != reflect.Ptr {
		args[0] = args[0].Elem()
	}
	res := h.fnValue.Call(args)[0]
	if res.IsNil() {
		return
	}
	// Type cast safe due to assertEventHandler()
	err = res.Interface().(error)
	glog.Error("Failed to handle event: ", err)
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`Internal server error`))
}
