FROM alpine:latest

RUN apk --no-cache add curl	
COPY main /	
COPY GeoLite2-Country.mmdb /	
ENTRYPOINT ["/main"]

