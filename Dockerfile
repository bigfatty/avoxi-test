FROM alpine:latest

RUN apk --no-cache add curl	
COPY main /	
ENTRYPOINT ["/main"]

