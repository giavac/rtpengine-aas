# rtpengine-aas
Go controller to interact with RTPEngine API

This application exposes two HTTP endpoints:
- `allocate_offer`
- `allocate_answer`

They can be used to generate an offer or answer SDP to be provided to an RTPEngine instance running locally.

This will cause RTPEngine to allocate the required resources.

It is a rudimentary Proof Of Concept.

## TODO

- Make HTTP listening port configurable
- Make RTPEngien listening socket configurable
- Run inside docker
- More options for the SDP manipulation
