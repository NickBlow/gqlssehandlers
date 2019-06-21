# Aims of this package

The aim of this package is to create a customisable set of handlers able to handle [GraphQL Subscriptions](https://graphql.github.io/graphql-spec/draft/#sec-Subscription) over Server Sent Events(SSE). While the spec does not specify a transport or serialisation mechanism, almost all subscription implementations are over WebSocket, and there's a dearth of libraries providing other transports. With Edge moving to a Chromium based engine, SSE are now supported by all the major browsers (and a polyfill for older browsers can be found [here](https://github.com/Yaffle/EventSource) - disclaimer: I haven't tried it out for all combinations of browsers), it seems to me that Server Sent Events are ready for prime time. 

Benefits over WebSockets are the simplicity of implementing - on the server it only requires setting a few headers and sending a streaming response, and on the client implementing an [Event Source](https://developer.mozilla.org/en-US/docs/Web/API/EventSource) is only a few lines of code. The other major benefit is the ability to stream over HTTP/2, which is not yet implemented for Web-Sockets, though it recently got a [proposed standard](https://tools.ietf.org/html/rfc8441).


# Further work

A definite value add would be the ability to create an [ApolloLink](https://github.com/apollographql/apollo-link), to let this be pluggable for people using Apollo Client.