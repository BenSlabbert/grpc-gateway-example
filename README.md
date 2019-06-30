# gRPC Gateway Example

# Set Up

Head over [here](https://github.com/grpc-ecosystem/grpc-gateway) and follow the set up

You basically need `protoc 3` installed

# Run

In the `apps` dir there are 3 programs to run:

- client - gRPC client
- server - gRPC server
- proxy - gRPC Rest proxy

There is also an accompanying Postman collection which consumes the gRPC services over REST

## Whats Next

- edit the `pb/messages.proto` file to add more message types
- edit the `pb/service.proto` file to add more services
