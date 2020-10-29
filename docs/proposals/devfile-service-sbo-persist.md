# Storing Service info and Link info in Devfile

## Abstract

## Motivation
Devfile should be complete description of an environment where the applications runs.
If application requires external service this information needs to be captured in Devfile. 

One of the odo goals is that if I clone project with Devfile and execute `odo push` I get working application.
If application requires external service (like db, cache, etc..), and this is not captured in Devfile, 
this won't work.
This is especially important when multiple developers collaborate on the same project. If information about the service and link is not in Devfile it means that the developers will have share this in other way.
If this happens odo and Devfile looses its purpose. 

## User Stories

### Story 1
As a developer working with odo want to get fully working application after cloning the project and running `odo push`.


### Story 2
As a developer working with odo, when I add new service to the application (like DB), I want to be able to easily share service definition and connection description (link) with my team so they can start application with the service in the same way as me. 


## Design overview
TODO

## Future evolution
TODO

