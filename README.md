# mao-bft
mao-bft is an efficient byzantine fault tolerance protocol without leader reelection

# Consensus layer
We design consensus layer independent from upper application layer, this mean that our BFT algorithm should be able to carry any message. To achieve this, we define the message to be any protobuf payload. In the mean time, each node in the cluster should be http server based on gRPC, we define RPC as below:

1. **SendToLeader**
This message sends requist it received locally to leader.
```protobuf
define RPC here
```

2. **AckTransaction**
This message is used by leader to ack to follower that TX has been checked in in system.
```protobuf
define RPC here
```

3. **PrepareValue**
This message is sent by leader and makes value known to every follower.
```protobuf
define RPC here
```

4. **EchoValue**
This message is send from everyone uppon receiving **PrepareValue**
```protobuf
define RPC here
```

5. **ReadyValue**
This message is send from everyone uppon receiving enough **EchoValue**
```protobuf
define RPC here
```
