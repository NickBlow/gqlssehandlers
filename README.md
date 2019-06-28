# Aims of this package

The aim of this package is to create a customisable set of handlers able to handle [GraphQL Subscriptions](https://graphql.github.io/graphql-spec/draft/#sec-Subscription) over Server Sent Events(SSE). While the spec does not specify a transport or serialisation mechanism, almost all subscription implementations are over WebSocket, and there's a dearth of libraries providing other transports. With Edge moving to a Chromium based engine, SSE are now supported by all the major browsers (and a polyfill for older browsers can be found [here](https://github.com/Yaffle/EventSource) - disclaimer: I haven't tried it out for all combinations of browsers), it seems to me that Server Sent Events are ready for prime time. 

Benefits over WebSockets are the simplicity of implementing - on the server it only requires setting a few headers and sending a streaming response, and on the client implementing an [Event Source](https://developer.mozilla.org/en-US/docs/Web/API/EventSource) is only a few lines of code. The other major benefit is the ability to stream over HTTP/2, which is not yet implemented for Web-Sockets, though it recently got a [proposed standard](https://tools.ietf.org/html/rfc8441).


# How to use

TBD

# Further work

A definite value add would be the ability to create an [ApolloLink](https://github.com/apollographql/apollo-link), to let this be pluggable for people using Apollo Client.

EventBuffer implementation -
If the EventBuffer size has been set, the server will store events in memory,
and re-send them if a client reconnects with a Last-Event-ID Header. This may result in duplicate messages being sent, so ensure your client is idempotent.
This will also only send events that the server itself received.
If you want to ensure a client always gets buffered messages, you can either use sticky sessions, route based on some hash, or multicast events to all servers.

Feedback from actual use to help improve DX, as well as examples for authentication etc.
