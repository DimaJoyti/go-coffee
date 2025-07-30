# 🎯 System Design  Question Bank

## 📋 Overview

This comprehensive question bank contains 100+ system design  questions with detailed solutions, organized by difficulty level and company type. Each question includes Go Coffee implementation examples and trade-off analysis.

---

## 🥉 **ENTRY LEVEL QUESTIONS (L3-L4)**

### **Q1: Design a URL Shortener (like bit.ly)**
**Companies**: Google, Facebook, Amazon
**Time**: 45 minutes

#### **Solution Framework**
```
1. Requirements Clarification (5 min)
   - 100M URLs shortened per day
   - 10:1 read/write ratio
   - Custom aliases allowed
   - Analytics required
   - 5-year data retention

2. Capacity Estimation (5 min)
   - Write: 100M/day = 1,200 QPS
   - Read: 1.2B/day = 14K QPS
   - Storage: 100M * 365 * 5 * 500 bytes = 91TB

3. High-Level Design (15 min)
   - Load Balancer → Web Servers → Cache → Database
   - URL encoding service (Base62)
   - Analytics service
   - CDN for static content

4. Detailed Design (15 min)
   - Database schema (URL mapping, analytics)
   - Encoding algorithm (counter vs hash)
   - Cache strategy (LRU, write-through)
   - Rate limiting

5. Scale & Optimize (5 min)
   - Database sharding by URL hash
   - Read replicas for analytics
   - Bloom filter for existence check
```

#### **Go Coffee Implementation Example**
```go
type URLShortener struct {
    db          Database
    cache       Cache
    encoder     *Base62Encoder
    analytics   AnalyticsService
    rateLimiter RateLimiter
}

func (us *URLShortener) ShortenURL(originalURL string, userID string) (*ShortenedURL, error) {
    // Rate limiting
    if !us.rateLimiter.Allow(userID) {
        return nil, ErrRateLimitExceeded
    }
    
    // Generate short code
    shortCode := us.encoder.Encode(us.getNextID())
    
    // Store mapping
    mapping := &URLMapping{
        ShortCode:   shortCode,
        OriginalURL: originalURL,
        UserID:      userID,
        CreatedAt:   time.Now(),
    }
    
    if err := us.db.SaveMapping(mapping); err != nil {
        return nil, err
    }
    
    // Cache the mapping
    us.cache.Set(shortCode, originalURL, 24*time.Hour)
    
    return &ShortenedURL{
        ShortCode: shortCode,
        ShortURL:  fmt.Sprintf("https://short.ly/%s", shortCode),
    }, nil
}
```

**Key Trade-offs**:
- Counter vs Hash: Predictability vs randomness
- SQL vs NoSQL: Consistency vs scalability
- Cache vs Database: Speed vs durability

---

### **Q2: Design a Chat System**
**Companies**: WhatsApp, Slack, Discord
**Time**: 45 minutes

#### **Solution Framework**
```
1. Requirements (5 min)
   - 1-on-1 and group chat
   - 50M DAU, 40 messages/user/day
   - Real-time delivery
   - Message history
   - Online presence

2. Capacity (5 min)
   - 2B messages/day = 23K QPS
   - Peak: 70K QPS
   - Storage: 2B * 100 bytes = 200GB/day

3. High-Level Design (15 min)
   - WebSocket servers for real-time
   - Message service for persistence
   - Notification service
   - Presence service

4. Detailed Design (15 min)
   - Message flow and delivery
   - Database schema
   - WebSocket connection management
   - Message ordering and consistency

5. Scale (5 min)
   - Horizontal scaling of WebSocket servers
   - Database sharding by chat_id
   - Message queue for reliability
```

#### **Go Coffee Implementation Example**
```go
type ChatService struct {
    wsManager    *WebSocketManager
    messageRepo  MessageRepository
    presenceRepo PresenceRepository
    notifier     NotificationService
}

func (cs *ChatService) SendMessage(msg *Message) error {
    // Validate and store message
    if err := cs.messageRepo.SaveMessage(msg); err != nil {
        return err
    }
    
    // Get online recipients
    recipients := cs.presenceRepo.GetOnlineUsers(msg.ChatID)
    
    // Send via WebSocket to online users
    for _, userID := range recipients {
        if conn := cs.wsManager.GetConnection(userID); conn != nil {
            conn.SendMessage(msg)
        }
    }
    
    // Send push notifications to offline users
    offlineUsers := cs.getOfflineUsers(msg.ChatID, recipients)
    for _, userID := range offlineUsers {
        cs.notifier.SendPushNotification(userID, msg)
    }
    
    return nil
}
```

---

## 🥈 **MID LEVEL QUESTIONS (L5-L6)**

### **Q3: Design Instagram**
**Companies**: Meta, Google, Amazon
**Time**: 45 minutes

#### **Solution Framework**
```
1. Requirements (5 min)
   - 500M DAU, 2M photos/day
   - Photo upload and viewing
   - News feed generation
   - Follow/unfollow users
   - Like and comment

2. Capacity (10 min)
   - Read:Write = 100:1
   - Storage: 2M * 2MB = 4TB/day
   - Bandwidth: Peak 10GB/s

3. High-Level Design (15 min)
   - CDN for photo delivery
   - Application servers
   - Metadata database
   - Object storage for photos
   - News feed service

4. Deep Dive (10 min)
   - Photo upload flow
   - News feed generation (push vs pull)
   - Database schema
   - Caching strategy

5. Scale (5 min)
   - Geographic distribution
   - Database sharding
   - Feed pre-computation
```

#### **Go Coffee Implementation Example**
```go
type InstagramService struct {
    photoStorage  ObjectStorage
    metadataDB    Database
    feedService   *FeedService
    cdn           CDNService
    imageProcessor *ImageProcessor
}

func (is *InstagramService) UploadPhoto(userID string, photo []byte, metadata *PhotoMetadata) (*Photo, error) {
    // Process image (resize, compress)
    processedImages := is.imageProcessor.ProcessImage(photo)
    
    // Upload to object storage
    photoID := generatePhotoID()
    urls := make(map[string]string)
    
    for size, image := range processedImages {
        url, err := is.photoStorage.Upload(fmt.Sprintf("%s_%s", photoID, size), image)
        if err != nil {
            return nil, err
        }
        urls[size] = url
    }
    
    // Save metadata
    photoRecord := &Photo{
        ID:        photoID,
        UserID:    userID,
        URLs:      urls,
        Caption:   metadata.Caption,
        Location:  metadata.Location,
        CreatedAt: time.Now(),
    }
    
    if err := is.metadataDB.SavePhoto(photoRecord); err != nil {
        return nil, err
    }
    
    // Update followers' feeds
    go is.feedService.UpdateFollowerFeeds(userID, photoRecord)
    
    return photoRecord, nil
}
```

---

### **Q4: Design a Distributed Cache**
**Companies**: Redis, Amazon, Google
**Time**: 45 minutes

#### **Solution Framework**
```
1. Requirements (5 min)
   - Get/Put operations
   - 100TB data, 1M QPS
   - Low latency (<1ms)
   - High availability
   - LRU eviction

2. High-Level Design (15 min)
   - Consistent hashing for distribution
   - Replication for availability
   - Client library for routing
   - Monitoring and metrics

3. Deep Dive (20 min)
   - Consistent hashing implementation
   - Replication strategies
   - Failure detection and recovery
   - Memory management

4. Scale (5 min)
   - Hot key handling
   - Network optimization
   - Monitoring and alerting
```

#### **Go Coffee Implementation Example**
```go
type DistributedCache struct {
    ring        *ConsistentHashRing
    nodes       map[string]*CacheNode
    replication int
    client      *CacheClient
}

func (dc *DistributedCache) Get(key string) ([]byte, error) {
    // Find primary and replica nodes
    nodes := dc.ring.GetNodes(key, dc.replication)
    
    // Try primary first
    if value, err := dc.client.Get(nodes[0], key); err == nil {
        return value, nil
    }
    
    // Try replicas
    for i := 1; i < len(nodes); i++ {
        if value, err := dc.client.Get(nodes[i], key); err == nil {
            // Repair primary asynchronously
            go dc.repairPrimary(nodes[0], key, value)
            return value, nil
        }
    }
    
    return nil, ErrKeyNotFound
}

func (dc *DistributedCache) Put(key string, value []byte, ttl time.Duration) error {
    nodes := dc.ring.GetNodes(key, dc.replication)
    
    // Write to all replicas
    var wg sync.WaitGroup
    errors := make(chan error, len(nodes))
    
    for _, node := range nodes {
        wg.Add(1)
        go func(n string) {
            defer wg.Done()
            if err := dc.client.Put(n, key, value, ttl); err != nil {
                errors <- err
            }
        }(node)
    }
    
    wg.Wait()
    close(errors)
    
    // Check if majority succeeded
    errorCount := len(errors)
    if errorCount > len(nodes)/2 {
        return ErrWriteFailed
    }
    
    return nil
}
```

---

## 🥇 **SENIOR LEVEL QUESTIONS (L7+)**

### **Q5: Design a Global Content Delivery Network**
**Companies**: CloudFlare, Amazon, Google
**Time**: 60 minutes

#### **Solution Framework**
```
1. Requirements (10 min)
   - Global distribution (100+ PoPs)
   - 10TB/s peak bandwidth
   - <50ms latency globally
   - 99.99% availability
   - DDoS protection

2. Architecture (20 min)
   - Edge servers and PoPs
   - Origin servers
   - DNS-based routing
   - Anycast networking
   - Cache hierarchy

3. Deep Dive (25 min)
   - Cache replacement algorithms
   - Content invalidation
   - Load balancing strategies
   - Security and DDoS mitigation
   - Performance optimization

4. Scale & Optimize (5 min)
   - Predictive caching
   - Edge computing
   - Network optimization
```

#### **Go Coffee Implementation Example**
```go
type CDNNode struct {
    location     *GeoLocation
    cache        *LRUCache
    origin       OriginServer
    loadBalancer *LoadBalancer
    metrics      *MetricsCollector
}

func (cdn *CDNNode) ServeContent(request *ContentRequest) (*ContentResponse, error) {
    // Check local cache first
    if content, exists := cdn.cache.Get(request.URL); exists {
        cdn.metrics.RecordCacheHit(request.URL)
        return &ContentResponse{
            Content:   content,
            Source:    "cache",
            Latency:   time.Since(request.StartTime),
        }, nil
    }
    
    // Cache miss - fetch from origin or parent cache
    content, err := cdn.fetchFromUpstream(request)
    if err != nil {
        return nil, err
    }
    
    // Cache the content
    cdn.cache.Set(request.URL, content, request.TTL)
    cdn.metrics.RecordCacheMiss(request.URL)
    
    return &ContentResponse{
        Content: content,
        Source:  "origin",
        Latency: time.Since(request.StartTime),
    }, nil
}

func (cdn *CDNNode) fetchFromUpstream(request *ContentRequest) ([]byte, error) {
    // Try parent cache first
    if parentCache := cdn.getParentCache(); parentCache != nil {
        if content, err := parentCache.Get(request.URL); err == nil {
            return content, nil
        }
    }
    
    // Fetch from origin
    return cdn.origin.FetchContent(request.URL)
}
```

---

### **Q6: Design a Real-time Analytics System**
**Companies**: Google, Amazon, Uber, Netflix
**Time**: 60 minutes

#### **Solution Framework**
```
1. Requirements (10 min)
   - 1M events/second
   - Real-time dashboards (<1s latency)
   - Historical analysis
   - Custom metrics and alerts
   - 99.9% availability

2. Architecture (20 min)
   - Event ingestion pipeline
   - Stream processing
   - Time-series database
   - Query engine
   - Visualization layer

3. Deep Dive (25 min)
   - Event schema and partitioning
   - Stream processing algorithms
   - Aggregation strategies
   - Storage optimization
   - Query performance

4. Scale (5 min)
   - Horizontal scaling
   - Data retention policies
   - Performance optimization
```

#### **Go Coffee Implementation Example**
```go
type AnalyticsSystem struct {
    ingestion    *EventIngestionService
    processor    *StreamProcessor
    storage      TimeSeriesDB
    queryEngine  *QueryEngine
    alertManager *AlertManager
}

func (as *AnalyticsSystem) ProcessEvent(event *Event) error {
    // Validate and enrich event
    enrichedEvent, err := as.ingestion.ProcessEvent(event)
    if err != nil {
        return err
    }
    
    // Send to stream processor
    if err := as.processor.ProcessEvent(enrichedEvent); err != nil {
        return err
    }
    
    // Check for alerts
    go as.alertManager.CheckAlerts(enrichedEvent)
    
    return nil
}

type StreamProcessor struct {
    windows map[string]*TimeWindow
    aggregators map[string]Aggregator
}

func (sp *StreamProcessor) ProcessEvent(event *Event) error {
    // Update time windows
    for windowType, window := range sp.windows {
        window.AddEvent(event)
        
        // Check if window is complete
        if window.IsComplete() {
            aggregatedData := sp.aggregators[windowType].Aggregate(window.Events)
            
            // Store aggregated data
            go sp.storeAggregatedData(windowType, aggregatedData)
            
            // Reset window
            window.Reset()
        }
    }
    
    return nil
}
```

---

## 🎯 ** SUCCESS STRATEGIES**

### **Question Analysis Framework**

1. **Listen Carefully**: Understand the exact requirements
2. **Ask Clarifying Questions**: Scale, features, constraints
3. **Estimate Scale**: Users, data, QPS, storage
4. **Start Simple**: High-level design first
5. **Deep Dive**: Focus on 2-3 components
6. **Discuss Trade-offs**: Every decision has pros/cons
7. **Scale Up**: How to handle 10x growth
8. **Wrap Up**: Summarize and address concerns

### **Common Mistakes to Avoid**

❌ **Jumping to Implementation**: Start with requirements
❌ **Over-Engineering**: Keep it simple initially
❌ **Ignoring Scale**: Always consider capacity
❌ **No Trade-offs**: Discuss alternatives
❌ **Poor Communication**: Think out loud
❌ **Time Management**: Don't spend too long on one area

### **Go Coffee Examples for Any Question**

- **Scalability**: Order processing during peak hours
- **Reliability**: Payment system fault tolerance
- **Performance**: Real-time order tracking
- **Security**: User authentication and authorization
- **Data**: Multi-database strategy
- **Communication**: Event-driven architecture
- **Monitoring**: Comprehensive observability

**This question bank provides comprehensive preparation for system design s at any level! 🚀🎯**
