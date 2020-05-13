# mao-bft
mao-bft is an efficient byzantine fault tolerance protocol without leader reelection

## Consensus layer
We design consensus layer independent from upper application layer, this mean that our BFT algorithm should be able to carry any message. To achieve this, we define the message to be any protobuf payload. In the mean time, each node in the cluster should be http server based on gRPC, we define RPC as below:

### Message Flow
We use RBC(Reliable Broadcast) as a building block. 
Optional: use threshold signature instead of "echo and prepare"

### Message Types
1. **AckTransaction**
This message is used by leader to ack to follower that TX has been checked in in system.
```protobuf
define RPC here
```

2. **PrepareValue**
This message is sent by leader and makes value known to every follower.
```protobuf
define RPC here
```

3. **EchoValue**
This message is send from everyone uppon receiving **PrepareValue**
```protobuf
define RPC here
```

4. **ReadyValue**
This message is send from everyone uppon receiving enough **EchoValue**
```protobuf
define RPC here
```


## Applications
We can utilize this protocal to implement different type of applications, in example folder we have a blockchain for demo