# disys-handin-4

Mandatory handin 4 for Distributed System class at ITU 2022.

Created by:

- Frederik Petersen (frepe)
- Andreas Severin Hauch Trøstrup (atro)
- Andreas Guldborg Hansen (aguh)

## How to run

## Expected results

## Notes

We attempted to automate the number of clients/peers being created and automatically finding an available port and connecting to new peers with the following (unsuccesful) algorithm:

1. Create a peer without port
1. Attempt to dial other ports, beginning at 5000, until an available one exists
1. Save this as peer.port
1. Listen to this port
1. Ping all other peers on ports 5000 up until but excluding own port
1. When getting pinged, create a connection to the port from the peer that called the function

Instead of spending a lot of time figuring out why this did not work, we instead decided to use a fixed number of peers at any given time (3) and where each program must be passed the --port flag of either 5000, 5001 or 5002, as described in [How to run](#how-to-run).

We noted subsequently, that the project description refers to the [Serf package]()
