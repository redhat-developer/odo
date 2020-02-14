# Event Listener Sink

This pod is created when an EventListener resource is created. The EventListener
sink pod listens for events. When it receives an event, the pod will create
resources based on the EventListener's TriggerTemplates, TriggerBindings, and
the event data.
