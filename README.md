# mao-bft
mao-bft is an efficient byzantine fault tolerance protocol without leader reelection

## Consensus layer
We design consensus layer independent from upper application layer, this mean that our BFT algorithm should be able to carry any message. To achieve this, we define the message to be any protobuf payload. In the mean time, each node in the cluster should be http server based on gRPC, we define RPC as below:

### Message Flow
We use RBC(Reliable Broadcast) as a building block. 
Optional: use threshold signature instead of "echo and prepare"

### Message Types
0. **Common Types**
This message defines common messages shared by RPC.
```protobuf
message Payload {
  string merkel_root;
  repeated string merkel_branch;
  bytes data;
}
```

1. **PrepareValue**
This message is sent by leader and makes value known to every follower.
```protobuf
rpc PrepareValue (PrepareValueRequest) returns (PrepareValueResponse) {}

message PrepareValueRequest {
  Payload payload;
}

message PrepareValueResponse {}
```

2. **EchoValue**
This message is send from everyone uppon receiving **PrepareValue**
```protobuf
rpc EchoValue (EchoValueRequest) returns (EchoValueResponse) {}

message EchoValueRequest {
  Payload payload;
}

message EchoValueResponse {}
```

3. **ReadyValue**
This message is send from everyone uppon receiving enough **EchoValue**
```protobuf
rpc EchoValue (EchoValueRequest) returns (EchoValueResponse) {}

message ReadyValueRequest {
  string merkel_root;
}

message ReadyValueResponse {}
```

### Application layer
1. **GetTransactionStatus**
This message is used by client to get TX status in the system.
```protobuf
define RPC here
```


## Applications
We can utilize this protocal to implement different type of applications, in example folder we have a blockchain for demo
