package plugin

import (
	api "github.com/metacosm/odo-event-api/odo/api/events"
)

var foo = "foo"

type ListenerRPCServer struct {
	// This is the real implementation
	Impl api.Listener
}

func (s *ListenerRPCServer) OnEvent(event api.Event, resp *string) error {
	resp = &foo
	return s.Impl.OnEvent(event)
}
func (s *ListenerRPCServer) OnAbort(abortError api.EventCausedAbortError, resp *string) error {
	resp = &foo
	s.Impl.OnAbort(abortError)
	return nil
}
func (s *ListenerRPCServer) Name(args interface{}, resp *string) error {
	name := s.Impl.Name()
	resp = &name
	return nil
}
