# webhook-demultiplexer

A request forwarder written in Golang. This applicaion receives an incoming request, based on the
provided configurations, this forwards the same request contents to multiple targets.

Consider you are integrating third party services like `https://www.textlocal.in/`. They provide only one web-hook url. But you want the web-hook to hit/notify three different servers/links. The only way this can be achieved is by a middleware code.

This code/app is that middleware that you exactly need.

## Getting Started

1. Copy the `config.example.json` => `config.json` and define your necessary configurations.
2. Start the app using `docker-compose build && docker-compose up` or `go run server.go`.
3. Now the app will be running on port `8080`. Write an `nginx` config to `proxy_pass` if needed.
