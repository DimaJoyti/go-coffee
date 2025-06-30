
package main

import (
    "go-coffee-ai-agents/internal/external/interfaces"
    "time"
)

func main() {
    // Test if Comment struct has the right fields
    _ = &interfaces.Comment{
        TaskID:    "test",
        UserID:    "user123", 
        Content:   "test content",
        CreatedAt: time.Now(),
        UpdatedAt: time.Now(),
    }
}
