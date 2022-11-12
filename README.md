# disys-handin-4

Mandatory handin 4 for Distributed System class at ITU 2022.

Created by:

- Frederik Petersen (frepe)
- Andreas Severin Hauch Trøstrup (atro)
- Andreas Guldborg Hansen (aguh)

## How to run

Run three terminals on the same computer. Run one of these commands in each terminal:

```sh
go run main.go --port 5000
```

```sh
go run main.go --port 5001
```

```sh
go run main.go --port 5002
```

Once all three clients/peers are running, in any terminal hit enter to request access.

## Expected results

Requesting access will log to a `peer_x.log` where `x` is a portnumber from 5000 to 5002 (both included).

What can be seen from the logs is that when a peer requests access, it will await responses from both other peers before accessing the restricted function. When access is granted, this can be seen as five `.` in the log.

Simultaneously, it will listen to requests and give access once it is done with the restricted function itself. If port 5002 requests access while peer on port 5000 is currently accessing the restricted function, it can be seen that the access is not granted to other peers before peer on port 5000 is finished accessing the restricted function. More details can be found in the report.

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
