# sf-challenge


## Build and Run

Run the development server:

```bash
RPC_ENDPOINT=<rpc_endpoint_with_slash_at_the_end/> go run main.go
```

It's important to include the slash at the end of the rpc endpoint!

## Explanation

This RESTful API application provides two endpoints:
- GET /blockreward/{slot}
- GET /syncduties/{slot}

### Blockreward Enpoint
A validator can receive a block from a MEV relay or build it internally. The reward of the validator is either the bid from a builder (last tx in block) if the block is received from a MEV relay or the transaction fees minus the base fees if the block is built internally in the validator node.

#### Implementation
- Missed slot vs slot in the future
  
  The blockreward handler checks if the slot is in the future if the block could not be fetched.

- Vanilla block vs MEV relay block
  
  The blockreward handler screens the transactions for a builder bid. It assumes that the fee recipient is the builder of the block.

### Syncduties Endpoint
Validators in a sync committee are fixed for about 27 hours. That's why it is possible to retrieve validator addresses with sync duties even for blocks that have not been built. If the slot is in the future the syncduties handler determines the corresponding epoch to retrieve the addresses of validators with sync committe duties.

### Frameworks

#### github.com/ethereum/go-ethereum/ethclient
Go-ethereum is used to connect to an execution node.

#### github.com/gin-gonic/gin
Gin is used to build the endpoints

## Examples

While the development server is running, you can use both endpoints.
```bash
curl -X 'GET' 'http://localhost:8080/blockreward/<slot>' -H 'accept: application/json'
curl -X 'GET' 'http://localhost:8080/syncduties/<slot>' -H 'accept: application/json'
```

### Blockreward Endpoint

Block produced by MEV relay
```bash
curl -X 'GET' 'http://localhost:8080/blockreward/9041486' -H 'accept: application/json'
```

Block built internally
```bash
curl -X 'GET' 'http://localhost:8080/blockreward/9041461' -H 'accept: application/json' 
```

Missed slot
```bash
curl -X 'GET' 'http://localhost:8080/blockreward/9041503' -H 'accept: application/json' 
```

### Syncduties Endpoint

Attention: It takes a while, so please be patient!
```bash
curl -X 'GET' 'http://localhost:8080/syncduties/9041503' -H 'accept: application/json' 
```
