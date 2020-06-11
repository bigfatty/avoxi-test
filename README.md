# Avoxi-Test

The grpc port is served on 8081.  I also created an http server which just uses the grpc service as the backend.

The http call is a POST request with a JSON body that contains the different whitelisted countries by field name.   I didn't put the time into making it a proper array as I hadn't worked with grpc before and I was running out of time due to issues at my current job.

``` 
curl http://localhost:8082/v1/46.40.128.15 \
--output -  \
-d '{"US":true, "AU":true, "CN": false}'
```

Moving forward, keeping the mmdb data and the whitelist data up to date would make sense.   This would involved a routine that pulls and updates the latest data automatically from MaxMind.   Also it would be nice to have some security on this so that we could make it easier for end users to maintain their lists rather than pass the full list each time.   Though we would need persistent storage to not cause confusion if the system goes down.

The data structures used for the mmdb search doesn't seem very efficient.  It might make more sense to pull the data into a local map where the response time might be faster.

Since this has grpc streaming, that would really be a much better solution rather than using REST.   Theoretically we could do something where the client just needs to upload their whitelist once and just send messages for each IP.   I don't have much experience with the protocol but theoretically that would make use of the strength of it.

There is a lot of additional work that goes beyond this test.   I'll just list them off to not be too verbose:

* timeouts
* retries
* auto-updates
* streaming
* alerts/subscriptions
* HA
* cache system
* security
* testing
* stat collections
* centralized persistent storage
