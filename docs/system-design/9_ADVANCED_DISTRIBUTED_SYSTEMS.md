# 🌐 9: Advanced Distributed Systems

## 📋 Overview

Master cutting-edge distributed systems patterns through Go Coffee's advanced architecture. This covers consensus algorithms, blockchain integration, edge computing, distributed AI systems, and next-generation architectural patterns.

## 🎯 Learning Objectives

By the end of this phase, you will:
- Implement consensus algorithms and distributed coordination
- Design blockchain integration and smart contract systems
- Master edge computing and IoT architectures
- Build distributed AI/ML systems at scale
- Understand next-generation distributed patterns
- Analyze Go Coffee's advanced distributed implementations

---

## 📖 9.1 Consensus Algorithms & Distributed Coordination

### Core Concepts

#### Consensus Algorithms
- **Raft**: Leader-based consensus for replicated state machines
- **PBFT**: Practical Byzantine Fault Tolerance
- **Paxos**: Classic consensus algorithm for distributed systems
- **CRDT**: Conflict-free Replicated Data Types
- **Vector Clocks**: Logical time in distributed systems

#### Distributed Coordination
- **Leader Election**: Selecting a coordinator in distributed systems
- **Distributed Locking**: Mutual exclusion across nodes
- **Membership Management**: Node discovery and failure detection
- **Configuration Management**: Distributed configuration consensus

#### CAP Theorem Applications
- **Consistency**: Strong vs eventual consistency trade-offs
- **Availability**: High availability patterns and techniques
- **Partition Tolerance**: Network partition handling strategies

### 🔍 Go Coffee Analysis

#### Study Raft Consensus Implementation

<augment_code_snippet path="internal/consensus/raft_node.go" mode="EXCERPT">
````go
type RaftNode struct {
    id          string
    state       NodeState
    currentTerm int64
    votedFor    string
    log         []LogEntry
    commitIndex int64
    lastApplied int64
    
    // Leader state
    nextIndex  map[string]int64
    matchIndex map[string]int64
    
    // Channels
    appendEntriesCh chan *AppendEntriesRequest
    requestVoteCh   chan *RequestVoteRequest
    commandCh       chan Command
    
    // Networking
    peers   map[string]RaftPeer
    storage Storage
    logger  *slog.Logger
    
    // Timers
    electionTimer  *time.Timer
    heartbeatTimer *time.Timer
    
    mutex sync.RWMutex
}

type NodeState int

const (
    Follower NodeState = iota
    Candidate
    Leader
)

type LogEntry struct {
    Term    int64       `json:"term"`
    Index   int64       `json:"index"`
    Command Command     `json:"command"`
    Type    EntryType   `json:"type"`
}

func (rn *RaftNode) Start() error {
    rn.mutex.Lock()
    rn.state = Follower
    rn.resetElectionTimer()
    rn.mutex.Unlock()
    
    go rn.run()
    
    rn.logger.Info("Raft node started", "id", rn.id)
    return nil
}

func (rn *RaftNode) run() {
    for {
        switch rn.getState() {
        case Follower:
            rn.runFollower()
        case Candidate:
            rn.runCandidate()
        case Leader:
            rn.runLeader()
        }
    }
}

func (rn *RaftNode) runFollower() {
    rn.logger.Debug("Running as follower", "term", rn.currentTerm)
    
    for rn.getState() == Follower {
        select {
        case req := <-rn.appendEntriesCh:
            rn.handleAppendEntries(req)
            
        case req := <-rn.requestVoteCh:
            rn.handleRequestVote(req)
            
        case <-rn.electionTimer.C:
            rn.logger.Info("Election timeout, becoming candidate")
            rn.becomeCandidate()
            
        case cmd := <-rn.commandCh:
            // Forward to leader if known
            if leader := rn.getLeader(); leader != "" {
                rn.forwardToLeader(leader, cmd)
            } else {
                cmd.ResponseCh <- CommandResponse{
                    Success: false,
                    Error:   "no leader available",
                }
            }
        }
    }
}

func (rn *RaftNode) runCandidate() {
    rn.logger.Info("Running as candidate", "term", rn.currentTerm)
    
    // Start election
    rn.mutex.Lock()
    rn.currentTerm++
    rn.votedFor = rn.id
    rn.resetElectionTimer()
    rn.mutex.Unlock()
    
    // Request votes from peers
    votes := rn.requestVotes()
    
    // Check if we won the election
    if votes > len(rn.peers)/2 {
        rn.becomeLeader()
        return
    }
    
    // Continue as candidate or step down
    for rn.getState() == Candidate {
        select {
        case req := <-rn.appendEntriesCh:
            if req.Term >= rn.currentTerm {
                rn.becomeFollower(req.Term)
                rn.handleAppendEntries(req)
            }
            
        case req := <-rn.requestVoteCh:
            rn.handleRequestVote(req)
            
        case <-rn.electionTimer.C:
            rn.logger.Info("Election timeout, starting new election")
            return // Restart election
        }
    }
}

func (rn *RaftNode) runLeader() {
    rn.logger.Info("Running as leader", "term", rn.currentTerm)
    
    // Initialize leader state
    rn.mutex.Lock()
    for peerID := range rn.peers {
        rn.nextIndex[peerID] = rn.getLastLogIndex() + 1
        rn.matchIndex[peerID] = 0
    }
    rn.mutex.Unlock()
    
    // Send initial heartbeats
    rn.sendHeartbeats()
    rn.resetHeartbeatTimer()
    
    for rn.getState() == Leader {
        select {
        case req := <-rn.appendEntriesCh:
            if req.Term > rn.currentTerm {
                rn.becomeFollower(req.Term)
                rn.handleAppendEntries(req)
            }
            
        case req := <-rn.requestVoteCh:
            rn.handleRequestVote(req)
            
        case cmd := <-rn.commandCh:
            rn.handleCommand(cmd)
            
        case <-rn.heartbeatTimer.C:
            rn.sendHeartbeats()
            rn.resetHeartbeatTimer()
        }
    }
}

func (rn *RaftNode) handleCommand(cmd Command) {
    rn.mutex.Lock()
    
    // Append to log
    entry := LogEntry{
        Term:    rn.currentTerm,
        Index:   rn.getLastLogIndex() + 1,
        Command: cmd,
        Type:    CommandEntry,
    }
    
    rn.log = append(rn.log, entry)
    rn.storage.StoreLogEntry(entry)
    
    rn.mutex.Unlock()
    
    // Replicate to followers
    rn.replicateEntry(entry, cmd.ResponseCh)
}

func (rn *RaftNode) replicateEntry(entry LogEntry, responseCh chan CommandResponse) {
    successCount := 1 // Leader counts as success
    
    var wg sync.WaitGroup
    var mutex sync.Mutex
    
    for peerID, peer := range rn.peers {
        wg.Add(1)
        go func(id string, p RaftPeer) {
            defer wg.Done()
            
            success := rn.replicateToPeer(id, p, entry)
            
            mutex.Lock()
            if success {
                successCount++
            }
            mutex.Unlock()
        }(peerID, peer)
    }
    
    wg.Wait()
    
    // Check if majority replicated
    if successCount > len(rn.peers)/2 {
        rn.commitEntry(entry.Index)
        responseCh <- CommandResponse{
            Success: true,
            Result:  "command committed",
        }
    } else {
        responseCh <- CommandResponse{
            Success: false,
            Error:   "failed to replicate to majority",
        }
    }
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 9.1: Implement Distributed Coordination

#### Step 1: Create Distributed Lock Service
```go
// internal/coordination/distributed_lock.go
package coordination

type DistributedLockService struct {
    raftNode    *RaftNode
    locks       map[string]*LockInfo
    lockTimeout time.Duration
    logger      *slog.Logger
    mutex       sync.RWMutex
}

type LockInfo struct {
    ID        string    `json:"id"`
    Owner     string    `json:"owner"`
    Resource  string    `json:"resource"`
    ExpiresAt time.Time `json:"expires_at"`
    CreatedAt time.Time `json:"created_at"`
}

type LockRequest struct {
    Resource string        `json:"resource"`
    Owner    string        `json:"owner"`
    TTL      time.Duration `json:"ttl"`
}

func NewDistributedLockService(raftNode *RaftNode) *DistributedLockService {
    return &DistributedLockService{
        raftNode:    raftNode,
        locks:       make(map[string]*LockInfo),
        lockTimeout: 30 * time.Second,
        logger:      slog.Default(),
    }
}

func (dls *DistributedLockService) AcquireLock(ctx context.Context, req *LockRequest) (*LockInfo, error) {
    // Create lock command
    lockID := uuid.New().String()
    command := Command{
        Type: "acquire_lock",
        Data: map[string]interface{}{
            "lock_id":  lockID,
            "resource": req.Resource,
            "owner":    req.Owner,
            "ttl":      req.TTL.Seconds(),
        },
        ResponseCh: make(chan CommandResponse, 1),
    }
    
    // Submit to Raft
    select {
    case dls.raftNode.commandCh <- command:
    case <-ctx.Done():
        return nil, ctx.Err()
    }
    
    // Wait for response
    select {
    case response := <-command.ResponseCh:
        if !response.Success {
            return nil, fmt.Errorf("failed to acquire lock: %s", response.Error)
        }
        
        lockInfo := &LockInfo{
            ID:        lockID,
            Owner:     req.Owner,
            Resource:  req.Resource,
            ExpiresAt: time.Now().Add(req.TTL),
            CreatedAt: time.Now(),
        }
        
        return lockInfo, nil
        
    case <-ctx.Done():
        return nil, ctx.Err()
    }
}

func (dls *DistributedLockService) ReleaseLock(ctx context.Context, lockID, owner string) error {
    command := Command{
        Type: "release_lock",
        Data: map[string]interface{}{
            "lock_id": lockID,
            "owner":   owner,
        },
        ResponseCh: make(chan CommandResponse, 1),
    }
    
    select {
    case dls.raftNode.commandCh <- command:
    case <-ctx.Done():
        return ctx.Err()
    }
    
    select {
    case response := <-command.ResponseCh:
        if !response.Success {
            return fmt.Errorf("failed to release lock: %s", response.Error)
        }
        return nil
        
    case <-ctx.Done():
        return ctx.Err()
    }
}

func (dls *DistributedLockService) ApplyCommand(cmd Command) CommandResponse {
    switch cmd.Type {
    case "acquire_lock":
        return dls.applyAcquireLock(cmd.Data)
    case "release_lock":
        return dls.applyReleaseLock(cmd.Data)
    default:
        return CommandResponse{
            Success: false,
            Error:   "unknown command type",
        }
    }
}

func (dls *DistributedLockService) applyAcquireLock(data map[string]interface{}) CommandResponse {
    dls.mutex.Lock()
    defer dls.mutex.Unlock()
    
    resource := data["resource"].(string)
    owner := data["owner"].(string)
    lockID := data["lock_id"].(string)
    ttl := time.Duration(data["ttl"].(float64)) * time.Second
    
    // Check if resource is already locked
    for _, lock := range dls.locks {
        if lock.Resource == resource && time.Now().Before(lock.ExpiresAt) {
            if lock.Owner != owner {
                return CommandResponse{
                    Success: false,
                    Error:   "resource already locked by another owner",
                }
            }
            // Extend existing lock
            lock.ExpiresAt = time.Now().Add(ttl)
            return CommandResponse{
                Success: true,
                Result:  "lock extended",
            }
        }
    }
    
    // Create new lock
    lock := &LockInfo{
        ID:        lockID,
        Owner:     owner,
        Resource:  resource,
        ExpiresAt: time.Now().Add(ttl),
        CreatedAt: time.Now(),
    }
    
    dls.locks[lockID] = lock
    
    dls.logger.Info("Lock acquired", 
        "lock_id", lockID,
        "resource", resource,
        "owner", owner)
    
    return CommandResponse{
        Success: true,
        Result:  "lock acquired",
    }
}

func (dls *DistributedLockService) StartLockCleanup() {
    ticker := time.NewTicker(10 * time.Second)
    go func() {
        for range ticker.C {
            dls.cleanupExpiredLocks()
        }
    }()
}

func (dls *DistributedLockService) cleanupExpiredLocks() {
    dls.mutex.Lock()
    defer dls.mutex.Unlock()
    
    now := time.Now()
    for lockID, lock := range dls.locks {
        if now.After(lock.ExpiresAt) {
            delete(dls.locks, lockID)
            dls.logger.Info("Lock expired and cleaned up", 
                "lock_id", lockID,
                "resource", lock.Resource)
        }
    }
}
```

#### Step 2: Implement CRDT for Conflict-Free Updates
```go
// internal/crdt/g_counter.go
package crdt

type GCounter struct {
    nodeID   string
    counters map[string]int64
    mutex    sync.RWMutex
}

func NewGCounter(nodeID string) *GCounter {
    return &GCounter{
        nodeID:   nodeID,
        counters: make(map[string]int64),
    }
}

func (gc *GCounter) Increment(delta int64) {
    gc.mutex.Lock()
    defer gc.mutex.Unlock()
    
    gc.counters[gc.nodeID] += delta
}

func (gc *GCounter) Value() int64 {
    gc.mutex.RLock()
    defer gc.mutex.RUnlock()
    
    var total int64
    for _, count := range gc.counters {
        total += count
    }
    return total
}

func (gc *GCounter) Merge(other *GCounter) {
    gc.mutex.Lock()
    defer gc.mutex.Unlock()
    
    for nodeID, count := range other.counters {
        if gc.counters[nodeID] < count {
            gc.counters[nodeID] = count
        }
    }
}

func (gc *GCounter) State() map[string]int64 {
    gc.mutex.RLock()
    defer gc.mutex.RUnlock()
    
    state := make(map[string]int64)
    for nodeID, count := range gc.counters {
        state[nodeID] = count
    }
    return state
}

// PN-Counter (Increment/Decrement Counter)
type PNCounter struct {
    positive *GCounter
    negative *GCounter
}

func NewPNCounter(nodeID string) *PNCounter {
    return &PNCounter{
        positive: NewGCounter(nodeID),
        negative: NewGCounter(nodeID),
    }
}

func (pn *PNCounter) Increment(delta int64) {
    if delta >= 0 {
        pn.positive.Increment(delta)
    } else {
        pn.negative.Increment(-delta)
    }
}

func (pn *PNCounter) Value() int64 {
    return pn.positive.Value() - pn.negative.Value()
}

func (pn *PNCounter) Merge(other *PNCounter) {
    pn.positive.Merge(other.positive)
    pn.negative.Merge(other.negative)
}
```

### 💡 Practice Question 9.1
**"Design a distributed coordination system for Go Coffee that handles leader election, distributed locking, and configuration management across multiple data centers with network partitions."**

**Solution Framework:**
1. **Consensus Implementation**
   - Raft consensus for leader election
   - Multi-Raft for horizontal scaling
   - Byzantine fault tolerance for untrusted environments
   - Hybrid consensus for different consistency requirements

2. **Distributed Coordination**
   - Hierarchical distributed locks
   - Lease-based coordination with renewal
   - Failure detection and recovery
   - Cross-datacenter coordination

3. **Partition Tolerance**
   - Split-brain prevention mechanisms
   - Quorum-based decision making
   - Graceful degradation strategies
   - Conflict resolution protocols

---

## 📖 9.2 Blockchain Integration & Smart Contracts

### Core Concepts

#### Blockchain Fundamentals
- **Consensus Mechanisms**: Proof of Work, Proof of Stake, Delegated PoS
- **Smart Contracts**: Self-executing contracts with code
- **Decentralized Applications (DApps)**: Blockchain-based applications
- **Oracles**: External data feeds for smart contracts
- **Layer 2 Solutions**: Scaling solutions like Lightning Network

#### Multi-Chain Architecture
- **Cross-Chain Communication**: Inter-blockchain protocols
- **Bridge Protocols**: Asset transfer between chains
- **Chain Abstraction**: Unified interface for multiple blockchains
- **Atomic Swaps**: Trustless cross-chain exchanges

### 🔍 Go Coffee Analysis

#### Study Multi-Chain Payment Integration

<augment_code_snippet path="crypto-wallet/internal/blockchain/multi_chain_client.go" mode="EXCERPT">
````go
type MultiChainClient struct {
    chains map[string]BlockchainClient
    config *MultiChainConfig
    logger *slog.Logger
}

type BlockchainClient interface {
    GetBalance(address string) (*big.Int, error)
    SendTransaction(tx *Transaction) (string, error)
    GetTransactionStatus(txHash string) (*TransactionStatus, error)
    EstimateGas(tx *Transaction) (*big.Int, error)
    GetBlockNumber() (uint64, error)
    SubscribeToBlocks(ch chan<- *Block) error
    SubscribeToLogs(filter *LogFilter, ch chan<- *Log) error
}

type EthereumClient struct {
    client   *ethclient.Client
    chainID  *big.Int
    gasPrice *big.Int
    logger   *slog.Logger
}

func (ec *EthereumClient) SendTransaction(tx *Transaction) (string, error) {
    // Convert to Ethereum transaction
    ethTx := &types.Transaction{
        To:       common.HexToAddress(tx.To),
        Value:    tx.Amount,
        Gas:      tx.GasLimit,
        GasPrice: tx.GasPrice,
        Data:     tx.Data,
        Nonce:    tx.Nonce,
    }
    
    // Sign transaction
    signedTx, err := ec.signTransaction(ethTx, tx.PrivateKey)
    if err != nil {
        return "", fmt.Errorf("failed to sign transaction: %w", err)
    }
    
    // Send transaction
    err = ec.client.SendTransaction(context.Background(), signedTx)
    if err != nil {
        return "", fmt.Errorf("failed to send transaction: %w", err)
    }
    
    txHash := signedTx.Hash().Hex()
    ec.logger.Info("Transaction sent", 
        "tx_hash", txHash,
        "to", tx.To,
        "amount", tx.Amount.String())
    
    return txHash, nil
}

type SmartContractManager struct {
    multiChain *MultiChainClient
    contracts  map[string]*ContractInfo
    logger     *slog.Logger
}

type ContractInfo struct {
    Address   string                 `json:"address"`
    ABI       abi.ABI               `json:"abi"`
    ChainID   string                `json:"chain_id"`
    Functions map[string]*Function  `json:"functions"`
}

func (scm *SmartContractManager) CallContract(chainID, contractAddr, method string, params []interface{}) (interface{}, error) {
    client, exists := scm.multiChain.chains[chainID]
    if !exists {
        return nil, fmt.Errorf("unsupported chain: %s", chainID)
    }
    
    contract, exists := scm.contracts[contractAddr]
    if !exists {
        return nil, fmt.Errorf("contract not found: %s", contractAddr)
    }
    
    // Prepare contract call
    callData, err := contract.ABI.Pack(method, params...)
    if err != nil {
        return nil, fmt.Errorf("failed to pack call data: %w", err)
    }
    
    // Execute call
    result, err := client.CallContract(&CallMsg{
        To:   contractAddr,
        Data: callData,
    })
    if err != nil {
        return nil, fmt.Errorf("contract call failed: %w", err)
    }
    
    // Unpack result
    unpacked, err := contract.ABI.Unpack(method, result)
    if err != nil {
        return nil, fmt.Errorf("failed to unpack result: %w", err)
    }
    
    return unpacked, nil
}

// Coffee Token Smart Contract Integration
type CoffeeTokenContract struct {
    scm         *SmartContractManager
    contractAddr string
    chainID     string
}

func (ctc *CoffeeTokenContract) MintTokens(to string, amount *big.Int) (string, error) {
    txHash, err := ctc.scm.SendTransaction(ctc.chainID, ctc.contractAddr, "mint", []interface{}{
        common.HexToAddress(to),
        amount,
    })
    if err != nil {
        return "", fmt.Errorf("failed to mint tokens: %w", err)
    }
    
    return txHash, nil
}

func (ctc *CoffeeTokenContract) ProcessPayment(from, to string, amount *big.Int, orderID string) (string, error) {
    // Call smart contract payment function
    txHash, err := ctc.scm.SendTransaction(ctc.chainID, ctc.contractAddr, "processPayment", []interface{}{
        common.HexToAddress(from),
        common.HexToAddress(to),
        amount,
        orderID,
    })
    if err != nil {
        return "", fmt.Errorf("failed to process payment: %w", err)
    }
    
    return txHash, nil
}

func (ctc *CoffeeTokenContract) GetLoyaltyPoints(userAddr string) (*big.Int, error) {
    result, err := ctc.scm.CallContract(ctc.chainID, ctc.contractAddr, "getLoyaltyPoints", []interface{}{
        common.HexToAddress(userAddr),
    })
    if err != nil {
        return nil, fmt.Errorf("failed to get loyalty points: %w", err)
    }
    
    points, ok := result.(*big.Int)
    if !ok {
        return nil, fmt.Errorf("invalid loyalty points result type")
    }
    
    return points, nil
}
````
</augment_code_snippet>

### 🛠️ Hands-on Exercise 9.2: Advanced Blockchain Integration

#### Step 1: Implement Cross-Chain Bridge
```go
// internal/blockchain/cross_chain_bridge.go
package blockchain

type CrossChainBridge struct {
    sourceChain      BlockchainClient
    destinationChain BlockchainClient
    bridgeContract   *SmartContract
    validator        *BridgeValidator
    logger           *slog.Logger
}

type BridgeTransaction struct {
    ID              string    `json:"id"`
    SourceChain     string    `json:"source_chain"`
    DestinationChain string   `json:"destination_chain"`
    SourceTxHash    string    `json:"source_tx_hash"`
    DestTxHash      string    `json:"dest_tx_hash"`
    Amount          *big.Int  `json:"amount"`
    Token           string    `json:"token"`
    Sender          string    `json:"sender"`
    Recipient       string    `json:"recipient"`
    Status          BridgeStatus `json:"status"`
    CreatedAt       time.Time `json:"created_at"`
    CompletedAt     *time.Time `json:"completed_at,omitempty"`
}

type BridgeStatus string

const (
    BridgeStatusPending   BridgeStatus = "pending"
    BridgeStatusValidated BridgeStatus = "validated"
    BridgeStatusCompleted BridgeStatus = "completed"
    BridgeStatusFailed    BridgeStatus = "failed"
)

func (ccb *CrossChainBridge) InitiateBridge(req *BridgeRequest) (*BridgeTransaction, error) {
    // Validate bridge request
    if err := ccb.validator.ValidateRequest(req); err != nil {
        return nil, fmt.Errorf("invalid bridge request: %w", err)
    }
    
    // Lock tokens on source chain
    lockTxHash, err := ccb.lockTokensOnSource(req)
    if err != nil {
        return nil, fmt.Errorf("failed to lock tokens: %w", err)
    }
    
    // Create bridge transaction record
    bridgeTx := &BridgeTransaction{
        ID:               uuid.New().String(),
        SourceChain:      req.SourceChain,
        DestinationChain: req.DestinationChain,
        SourceTxHash:     lockTxHash,
        Amount:           req.Amount,
        Token:            req.Token,
        Sender:           req.Sender,
        Recipient:        req.Recipient,
        Status:           BridgeStatusPending,
        CreatedAt:        time.Now(),
    }
    
    // Start validation process
    go ccb.processBridgeTransaction(bridgeTx)
    
    return bridgeTx, nil
}

func (ccb *CrossChainBridge) processBridgeTransaction(bridgeTx *BridgeTransaction) {
    // Wait for source transaction confirmation
    confirmed, err := ccb.waitForConfirmation(bridgeTx.SourceChain, bridgeTx.SourceTxHash)
    if err != nil || !confirmed {
        bridgeTx.Status = BridgeStatusFailed
        ccb.logger.Error("Source transaction failed", "bridge_id", bridgeTx.ID, "error", err)
        return
    }
    
    // Validate with multiple validators
    validated, err := ccb.validator.ValidateTransaction(bridgeTx)
    if err != nil || !validated {
        bridgeTx.Status = BridgeStatusFailed
        ccb.logger.Error("Bridge validation failed", "bridge_id", bridgeTx.ID, "error", err)
        return
    }
    
    bridgeTx.Status = BridgeStatusValidated
    
    // Mint tokens on destination chain
    destTxHash, err := ccb.mintTokensOnDestination(bridgeTx)
    if err != nil {
        bridgeTx.Status = BridgeStatusFailed
        ccb.logger.Error("Failed to mint on destination", "bridge_id", bridgeTx.ID, "error", err)
        return
    }
    
    bridgeTx.DestTxHash = destTxHash
    bridgeTx.Status = BridgeStatusCompleted
    now := time.Now()
    bridgeTx.CompletedAt = &now
    
    ccb.logger.Info("Bridge transaction completed", 
        "bridge_id", bridgeTx.ID,
        "source_tx", bridgeTx.SourceTxHash,
        "dest_tx", bridgeTx.DestTxHash)
}

func (ccb *CrossChainBridge) lockTokensOnSource(req *BridgeRequest) (string, error) {
    // Call bridge contract to lock tokens
    tx := &Transaction{
        To:       ccb.bridgeContract.Address,
        Amount:   req.Amount,
        GasLimit: 100000,
        Data:     ccb.encodeLockCall(req),
    }
    
    return ccb.sourceChain.SendTransaction(tx)
}

func (ccb *CrossChainBridge) mintTokensOnDestination(bridgeTx *BridgeTransaction) (string, error) {
    // Call destination bridge contract to mint tokens
    tx := &Transaction{
        To:       ccb.bridgeContract.Address,
        Amount:   big.NewInt(0),
        GasLimit: 150000,
        Data:     ccb.encodeMintCall(bridgeTx),
    }
    
    return ccb.destinationChain.SendTransaction(tx)
}
```

#### Step 2: Implement Oracle Network
```go
// internal/blockchain/oracle_network.go
package blockchain

type OracleNetwork struct {
    oracles    map[string]*Oracle
    aggregator *PriceAggregator
    validator  *DataValidator
    logger     *slog.Logger
}

type Oracle struct {
    ID          string                 `json:"id"`
    Name        string                 `json:"name"`
    Endpoint    string                 `json:"endpoint"`
    Reputation  float64               `json:"reputation"`
    DataSources []string              `json:"data_sources"`
    LastUpdate  time.Time             `json:"last_update"`
    IsActive    bool                  `json:"is_active"`
}

type PriceData struct {
    Symbol    string    `json:"symbol"`
    Price     *big.Int  `json:"price"`
    Timestamp time.Time `json:"timestamp"`
    Source    string    `json:"source"`
    Signature string    `json:"signature"`
}

func (on *OracleNetwork) GetAggregatedPrice(symbol string) (*PriceData, error) {
    // Collect prices from multiple oracles
    var prices []*PriceData
    var wg sync.WaitGroup
    var mutex sync.Mutex
    
    for _, oracle := range on.oracles {
        if !oracle.IsActive {
            continue
        }
        
        wg.Add(1)
        go func(o *Oracle) {
            defer wg.Done()
            
            price, err := on.fetchPriceFromOracle(o, symbol)
            if err != nil {
                on.logger.Error("Failed to fetch price from oracle", 
                    "oracle", o.Name, "symbol", symbol, "error", err)
                return
            }
            
            mutex.Lock()
            prices = append(prices, price)
            mutex.Unlock()
        }(oracle)
    }
    
    wg.Wait()
    
    if len(prices) == 0 {
        return nil, fmt.Errorf("no price data available for %s", symbol)
    }
    
    // Aggregate prices using weighted median
    aggregatedPrice := on.aggregator.WeightedMedian(prices)
    
    // Validate aggregated price
    if err := on.validator.ValidatePrice(aggregatedPrice); err != nil {
        return nil, fmt.Errorf("price validation failed: %w", err)
    }
    
    return aggregatedPrice, nil
}

func (on *OracleNetwork) UpdateSmartContractPrice(symbol string, price *PriceData) error {
    // Update price in smart contract
    contractCall := &ContractCall{
        Contract: "PriceOracle",
        Method:   "updatePrice",
        Params: []interface{}{
            symbol,
            price.Price,
            price.Timestamp.Unix(),
            price.Signature,
        },
    }
    
    txHash, err := on.executeContractCall(contractCall)
    if err != nil {
        return fmt.Errorf("failed to update contract price: %w", err)
    }
    
    on.logger.Info("Price updated in smart contract", 
        "symbol", symbol,
        "price", price.Price.String(),
        "tx_hash", txHash)
    
    return nil
}
```

### 💡 Practice Question 9.2
**"Design a blockchain integration for Go Coffee that supports multi-chain payments, loyalty tokens, and supply chain tracking with oracle price feeds and cross-chain interoperability."**

**Solution Framework:**
1. **Multi-Chain Architecture**
   - Support for Ethereum, Polygon, BSC, and Solana
   - Cross-chain bridge for asset transfers
   - Unified wallet interface
   - Chain-specific optimizations

2. **Smart Contract System**
   - Payment processing contracts
   - Loyalty token contracts
   - Supply chain tracking contracts
   - Oracle integration contracts

3. **Interoperability Solutions**
   - Cross-chain communication protocols
   - Atomic swap mechanisms
   - Bridge security and validation
   - Oracle network for external data

---

## 📖 9.3 Edge Computing & IoT Integration

### Core Concepts

#### Edge Computing Architecture
- **Edge Nodes**: Distributed computing at network edge
- **Fog Computing**: Intermediate layer between edge and cloud
- **Mobile Edge Computing**: 5G-enabled edge processing
- **Content Delivery**: Edge-based content caching and delivery

#### IoT Integration Patterns
- **Device Management**: Registration, configuration, monitoring
- **Data Collection**: Sensor data aggregation and processing
- **Real-time Processing**: Stream processing at the edge
- **Offline Capability**: Local processing when disconnected

### 🔍 Go Coffee Analysis

#### Study IoT Coffee Machine Integration

<augment_code_snippet path="iot-integration/internal/device/coffee_machine.go" mode="EXCERPT">
````go
type CoffeeMachine struct {
    ID           string                 `json:"id"`
    Name         string                 `json:"name"`
    Location     *Location             `json:"location"`
    Status       MachineStatus         `json:"status"`
    Capabilities []MachineCapability   `json:"capabilities"`
    Sensors      map[string]*Sensor    `json:"sensors"`
    LastSeen     time.Time             `json:"last_seen"`
    
    // Edge processing
    edgeProcessor *EdgeProcessor
    dataBuffer    *CircularBuffer
    
    // Communication
    mqttClient    mqtt.Client
    httpClient    *http.Client
    
    logger *slog.Logger
}

type MachineStatus string

const (
    StatusOnline     MachineStatus = "online"
    StatusOffline    MachineStatus = "offline"
    StatusMaintenance MachineStatus = "maintenance"
    StatusError      MachineStatus = "error"
)

type Sensor struct {
    ID       string      `json:"id"`
    Type     SensorType  `json:"type"`
    Value    interface{} `json:"value"`
    Unit     string      `json:"unit"`
    LastRead time.Time   `json:"last_read"`
}

type SensorType string

const (
    SensorTemperature    SensorType = "temperature"
    SensorPressure       SensorType = "pressure"
    SensorWaterLevel     SensorType = "water_level"
    SensorBeanLevel      SensorType = "bean_level"
    SensorMilkLevel      SensorType = "milk_level"
    SensorCleaningStatus SensorType = "cleaning_status"
)

func (cm *CoffeeMachine) Start() error {
    // Initialize MQTT connection
    if err := cm.connectMQTT(); err != nil {
        return fmt.Errorf("failed to connect MQTT: %w", err)
    }
    
    // Start sensor monitoring
    go cm.monitorSensors()
    
    // Start edge processing
    go cm.processDataAtEdge()
    
    // Start heartbeat
    go cm.sendHeartbeat()
    
    cm.logger.Info("Coffee machine started", "id", cm.ID, "location", cm.Location)
    return nil
}

func (cm *CoffeeMachine) monitorSensors() {
    ticker := time.NewTicker(5 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        for sensorID, sensor := range cm.Sensors {
            value, err := cm.readSensorValue(sensor)
            if err != nil {
                cm.logger.Error("Failed to read sensor", "sensor_id", sensorID, "error", err)
                continue
            }
            
            sensor.Value = value
            sensor.LastRead = time.Now()
            
            // Add to data buffer for edge processing
            dataPoint := &SensorDataPoint{
                MachineID: cm.ID,
                SensorID:  sensorID,
                Value:     value,
                Timestamp: time.Now(),
            }
            
            cm.dataBuffer.Add(dataPoint)
            
            // Check for alerts
            if alert := cm.checkSensorAlert(sensor); alert != nil {
                cm.handleAlert(alert)
            }
        }
    }
}

func (cm *CoffeeMachine) processDataAtEdge() {
    for {
        // Process buffered data
        dataPoints := cm.dataBuffer.GetBatch(100)
        if len(dataPoints) == 0 {
            time.Sleep(1 * time.Second)
            continue
        }
        
        // Perform edge analytics
        analytics := cm.edgeProcessor.ProcessBatch(dataPoints)
        
        // Send analytics to cloud if significant
        if analytics.IsSignificant() {
            cm.sendAnalyticsToCloud(analytics)
        }
        
        // Local decision making
        if decision := cm.makeLocalDecision(analytics); decision != nil {
            cm.executeLocalDecision(decision)
        }
    }
}

func (cm *CoffeeMachine) ProcessOrder(order *Order) (*OrderResult, error) {
    // Check machine capabilities
    if !cm.canProcessOrder(order) {
        return nil, fmt.Errorf("machine cannot process this order")
    }
    
    // Check resource levels
    if err := cm.checkResourceLevels(order); err != nil {
        return nil, fmt.Errorf("insufficient resources: %w", err)
    }
    
    // Start brewing process
    brewingProcess := &BrewingProcess{
        OrderID:   order.ID,
        Recipe:    order.Recipe,
        StartTime: time.Now(),
        Status:    BrewingStatusStarted,
    }
    
    // Execute brewing steps
    result, err := cm.executeBrewing(brewingProcess)
    if err != nil {
        return nil, fmt.Errorf("brewing failed: %w", err)
    }
    
    // Update machine state
    cm.updateMachineState(result)
    
    // Send completion notification
    cm.notifyOrderCompletion(order.ID, result)
    
    return result, nil
}

func (cm *CoffeeMachine) executeBrewing(process *BrewingProcess) (*OrderResult, error) {
    steps := []BrewingStep{
        {Name: "grind_beans", Duration: 10 * time.Second},
        {Name: "heat_water", Duration: 30 * time.Second},
        {Name: "extract_coffee", Duration: 25 * time.Second},
        {Name: "add_milk", Duration: 15 * time.Second},
        {Name: "finalize", Duration: 5 * time.Second},
    }
    
    for _, step := range steps {
        cm.logger.Info("Executing brewing step", 
            "order_id", process.OrderID,
            "step", step.Name)
        
        if err := cm.executeBrewingStep(step); err != nil {
            return nil, fmt.Errorf("step %s failed: %w", step.Name, err)
        }
        
        // Send progress update
        cm.sendProgressUpdate(process.OrderID, step.Name)
    }
    
    result := &OrderResult{
        OrderID:     process.OrderID,
        Status:      OrderStatusCompleted,
        CompletedAt: time.Now(),
        Quality:     cm.assessQuality(),
    }
    
    return result, nil
}
````
</augment_code_snippet>

### 💡 Practice Question 9.3
**"Design an edge computing architecture for Go Coffee that supports 10,000+ IoT coffee machines with real-time processing, offline capability, and intelligent decision making at the edge."**

**Solution Framework:**
1. **Edge Architecture**
   - Hierarchical edge nodes (device → edge → fog → cloud)
   - Local processing capabilities
   - Intelligent data filtering and aggregation
   - Offline operation with sync capabilities

2. **IoT Device Management**
   - Automated device registration and provisioning
   - Over-the-air updates and configuration
   - Health monitoring and predictive maintenance
   - Security and authentication

3. **Real-time Processing**
   - Stream processing at edge nodes
   - Local decision making algorithms
   - Event-driven architecture
   - Low-latency response requirements

---

## 🎯 9 Completion Checklist

### Knowledge Mastery
- [ ] Understand consensus algorithms and distributed coordination
- [ ] Can design blockchain integration and smart contract systems
- [ ] Know edge computing and IoT architecture patterns
- [ ] Understand distributed AI/ML systems
- [ ] Can implement next-generation distributed patterns

### Practical Skills
- [ ] Can implement Raft consensus and distributed locks
- [ ] Can build multi-chain blockchain integrations
- [ ] Can design edge computing architectures
- [ ] Can create distributed AI systems
- [ ] Can handle advanced distributed system challenges

### Go Coffee Analysis
- [ ] Analyzed consensus and coordination implementations
- [ ] Studied blockchain and smart contract integrations
- [ ] Examined edge computing and IoT patterns
- [ ] Understood distributed AI architectures
- [ ] Identified next-generation optimization opportunities

###  Readiness
- [ ] Can design consensus systems for distributed coordination
- [ ] Can explain blockchain integration trade-offs
- [ ] Can implement edge computing solutions
- [ ] Can handle distributed AI system scenarios
- [ ] Can discuss next-generation distributed patterns

---

## 🚀 Next Steps

Ready for **10:  Mastery & Certification**:
- Advanced practice scenarios and mock s
- Comprehensive final assessment
- Bronze/Silver/Gold certification achievement
- Career advancement preparation
- System design mastery validation

**Excellent progress on mastering advanced distributed systems! 🎉**
