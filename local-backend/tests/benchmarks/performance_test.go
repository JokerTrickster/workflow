package benchmarks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JokerTrickster/workflow/local-backend/internal/domain"
	"github.com/JokerTrickster/workflow/local-backend/internal/infrastructure"
	"github.com/JokerTrickster/workflow/local-backend/internal/usecase"
	"github.com/stretchr/testify/require"
)

// setupBenchmarkDatabase creates a temporary database for benchmarking
func setupBenchmarkDatabase(b *testing.B) (*infrastructure.SQLiteRepository, func()) {
	tempDir, err := os.MkdirTemp("", "benchmark-*")
	require.NoError(b, err)

	dbPath := filepath.Join(tempDir, "benchmark.db")
	
	config := &infrastructure.DatabaseConfig{
		Driver:       "sqlite",
		DSN:          dbPath,
		MaxIdleConns: 10,
		MaxOpenConns: 20,
	}

	logConfig := &infrastructure.LoggingConfig{
		Level:  "error", // Minimal logging for benchmarks
		Format: "text",
		Output: "stdout",
	}
	logger := infrastructure.NewLogger(logConfig)

	repo, err := infrastructure.NewSQLiteRepository(config, logger)
	require.NoError(b, err)

	err = repo.Initialize()
	require.NoError(b, err)

	cleanup := func() {
		repo.Close()
		os.RemoveAll(tempDir)
	}

	return repo, cleanup
}

// BenchmarkClaudeService for benchmarks
type BenchmarkClaudeService struct{}

func (m *BenchmarkClaudeService) ProcessRequest(ctx context.Context, message *domain.Message) (string, error) {
	// Simulate realistic processing time (10-100ms)
	time.Sleep(50 * time.Millisecond)
	
	payload := message.Payload
	if payload == nil {
		return "Empty request processed.", nil
	}

	code, _ := payload["code"].(string)
	task, _ := payload["task"].(string)

	return fmt.Sprintf("Analysis complete for task: %s. Code length: %d characters.", task, len(code)), nil
}

// BenchmarkMessageCreation tests domain entity creation performance
func BenchmarkMessageCreation(b *testing.B) {
	payload := map[string]interface{}{
		"code": "func main() { println(\"hello\") }",
		"task": "analyze this code",
		"metadata": map[string]interface{}{
			"priority": "high",
			"timeout":  30,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		message := domain.NewMessage(
			fmt.Sprintf("msg-%d", i),
			domain.MessageTypeWorkRequest,
			payload,
		)
		message.SetContextID(fmt.Sprintf("ctx-%d", i%100)) // Reuse contexts
		
		// Prevent compiler optimization
		_ = message.ID
	}
}

// BenchmarkRequestCreation tests request entity creation
func BenchmarkRequestCreation(b *testing.B) {
	requestData := `{
		"code": "func benchmark() { for i := 0; i < 1000; i++ { println(i) } }",
		"task": "optimize this loop for performance"
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := domain.NewRequest(
			fmt.Sprintf("req-%d", i),
			fmt.Sprintf("msg-%d", i),
			fmt.Sprintf("ctx-%d", i%50), // Reuse contexts
			requestData,
		)
		
		// Simulate request lifecycle
		request.Start()
		request.Complete("Optimization suggestions provided")
		
		_ = request.Status
	}
}

// BenchmarkDatabaseOperations tests database performance
func BenchmarkDatabaseOperations(b *testing.B) {
	repo, cleanup := setupBenchmarkDatabase(b)
	defer cleanup()

	ctx := context.Background()

	// Pre-create some test data
	messages := make([]*domain.Message, 1000)
	for i := 0; i < 1000; i++ {
		messages[i] = domain.NewMessage(
			fmt.Sprintf("bench-msg-%d", i),
			domain.MessageTypeWorkRequest,
			map[string]interface{}{
				"code": fmt.Sprintf("func test%d() {}", i),
				"task": "benchmark test",
			},
		)
		messages[i].SetContextID(fmt.Sprintf("ctx-%d", i%10))
	}

	b.Run("MessageSave", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			message := messages[i%1000]
			message.ID = fmt.Sprintf("save-msg-%d", i) // Unique ID for each iteration
			err := repo.SaveMessage(ctx, message)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Save some messages for retrieval benchmarks
	for i := 0; i < 100; i++ {
		msg := messages[i]
		err := repo.SaveMessage(ctx, msg)
		require.NoError(b, err)
	}

	b.Run("MessageRetrieve", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			id := fmt.Sprintf("bench-msg-%d", i%100)
			message, err := repo.GetMessageByID(ctx, id)
			if err != nil {
				b.Fatal(err)
			}
			_ = message
		}
	})

	b.Run("MessagesByContext", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			contextID := fmt.Sprintf("ctx-%d", i%10)
			messages, err := repo.GetMessagesByContextID(ctx, contextID)
			if err != nil {
				b.Fatal(err)
			}
			_ = messages
		}
	})

	// Prepare requests for benchmarking
	requests := make([]*domain.Request, 100)
	for i := 0; i < 100; i++ {
		requests[i] = domain.NewRequest(
			fmt.Sprintf("bench-req-%d", i),
			fmt.Sprintf("bench-msg-%d", i),
			fmt.Sprintf("ctx-%d", i%5),
			fmt.Sprintf(`{"code": "func test%d() {}", "task": "benchmark"}`, i),
		)
		err := repo.SaveRequest(ctx, requests[i])
		require.NoError(b, err)
	}

	b.Run("RequestRetrieve", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			id := fmt.Sprintf("bench-req-%d", i%100)
			request, err := repo.GetRequestByID(ctx, id)
			if err != nil {
				b.Fatal(err)
			}
			_ = request
		}
	})

	b.Run("RequestUpdate", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			request := requests[i%100]
			if request.Status == domain.RequestStatusPending {
				request.Start()
			} else if request.Status == domain.RequestStatusProcessing {
				request.Complete("Benchmark complete")
			} else {
				request.Status = domain.RequestStatusPending // Reset for next iteration
			}
			
			err := repo.UpdateRequest(ctx, request)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkMessageProcessing tests end-to-end message processing
func BenchmarkMessageProcessing(b *testing.B) {
	repo, cleanup := setupBenchmarkDatabase(b)
	defer cleanup()

	logConfig := &infrastructure.LoggingConfig{
		Level:  "error",
		Format: "text",
		Output: "stdout",
	}
	logger := infrastructure.NewLogger(logConfig)

	claudeService := &BenchmarkClaudeService{}
	requestService := usecase.NewRequestService(repo, claudeService, logger)
	messageProcessor := usecase.NewMessageProcessor(repo, repo, requestService, logger)

	ctx := context.Background()

	// Benchmark single message processing
	b.Run("SingleMessage", func(b *testing.B) {
		messages := make([]*domain.Message, b.N)
		for i := 0; i < b.N; i++ {
			messages[i] = domain.NewMessage(
				fmt.Sprintf("proc-msg-%d", i),
				domain.MessageTypeWorkRequest,
				map[string]interface{}{
					"code": fmt.Sprintf("func processTest%d() { return %d }", i, i*2),
					"task": "analyze and optimize",
				},
			)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := messageProcessor.ProcessMessage(ctx, messages[i])
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	// Benchmark message processing with context
	b.Run("MessageWithContext", func(b *testing.B) {
		messages := make([]*domain.Message, b.N)
		for i := 0; i < b.N; i++ {
			messages[i] = domain.NewMessage(
				fmt.Sprintf("ctx-msg-%d", i),
				domain.MessageTypeWorkRequest,
				map[string]interface{}{
					"code": fmt.Sprintf("func contextTest%d() {}", i),
					"task": "context processing test",
				},
			)
			// Distribute across multiple contexts
			messages[i].SetContextID(fmt.Sprintf("bench-ctx-%d", i%10))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			err := messageProcessor.ProcessMessage(ctx, messages[i])
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// BenchmarkConcurrentProcessing tests concurrent processing performance
func BenchmarkConcurrentProcessing(b *testing.B) {
	repo, cleanup := setupBenchmarkDatabase(b)
	defer cleanup()

	logConfig := &infrastructure.LoggingConfig{
		Level:  "error",
		Format: "text",
		Output: "stdout",
	}
	logger := infrastructure.NewLogger(logConfig)

	claudeService := &BenchmarkClaudeService{}
	requestService := usecase.NewRequestService(repo, claudeService, logger)
	messageProcessor := usecase.NewMessageProcessor(repo, repo, requestService, logger)

	ctx := context.Background()

	b.Run("Concurrent10", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				message := domain.NewMessage(
					fmt.Sprintf("concurrent-msg-%d-%d", b.N, i),
					domain.MessageTypeWorkRequest,
					map[string]interface{}{
						"code": fmt.Sprintf("func concurrentTest%d() {}", i),
						"task": "concurrent processing test",
					},
				)
				
				err := messageProcessor.ProcessMessage(ctx, message)
				if err != nil {
					b.Fatal(err)
				}
				i++
			}
		})
	})
}

// BenchmarkContextOperations tests processing context performance
func BenchmarkContextOperations(b *testing.B) {
	repo, cleanup := setupBenchmarkDatabase(b)
	defer cleanup()

	ctx := context.Background()

	// Pre-create contexts
	contexts := make([]*domain.ProcessingContext, 100)
	for i := 0; i < 100; i++ {
		contexts[i] = domain.NewProcessingContext(fmt.Sprintf("bench-ctx-%d", i))
		
		// Add some messages to make it realistic
		contexts[i].AddUserMessage(fmt.Sprintf("User message %d", i))
		contexts[i].AddAssistantMessage(fmt.Sprintf("Assistant response %d", i))
		contexts[i].SetMetadata("request_count", fmt.Sprintf("%d", i))
		
		err := repo.SaveProcessingContext(ctx, contexts[i])
		require.NoError(b, err)
	}

	b.Run("ContextSave", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			context := domain.NewProcessingContext(fmt.Sprintf("save-ctx-%d", i))
			context.AddUserMessage("Benchmark user message")
			context.AddAssistantMessage("Benchmark assistant response")
			
			err := repo.SaveProcessingContext(ctx, context)
			if err != nil {
				b.Fatal(err)
			}
		}
	})

	b.Run("ContextRetrieve", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			id := fmt.Sprintf("bench-ctx-%d", i%100)
			context, err := repo.GetProcessingContextByID(ctx, id)
			if err != nil {
				b.Fatal(err)
			}
			_ = context
		}
	})

	b.Run("ContextExpiredQuery", func(b *testing.B) {
		maxAge := 1 * time.Hour
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			contexts, err := repo.GetExpiredProcessingContexts(ctx, maxAge)
			if err != nil {
				b.Fatal(err)
			}
			_ = contexts
		}
	})
}

// BenchmarkMemoryUsage tests memory efficiency
func BenchmarkMemoryUsage(b *testing.B) {
	b.Run("MessageMemoryFootprint", func(b *testing.B) {
		messages := make([]*domain.Message, b.N)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			messages[i] = domain.NewMessage(
				fmt.Sprintf("mem-msg-%d", i),
				domain.MessageTypeWorkRequest,
				map[string]interface{}{
					"code": "func memoryTest() { println(\"memory test\") }",
					"task": "memory usage benchmark",
					"large_data": make([]byte, 1024), // 1KB of data
				},
			)
		}
		
		// Keep messages in memory to measure retention
		_ = messages
	})

	b.Run("RequestMemoryFootprint", func(b *testing.B) {
		requests := make([]*domain.Request, b.N)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			requests[i] = domain.NewRequest(
				fmt.Sprintf("mem-req-%d", i),
				fmt.Sprintf("mem-msg-%d", i),
				fmt.Sprintf("mem-ctx-%d", i%10),
				fmt.Sprintf(`{
					"code": "func memoryRequest%d() { /* large function */ }",
					"task": "analyze memory usage",
					"data": "%s"
				}`, i, string(make([]byte, 512))), // 512 bytes
			)
		}
		
		_ = requests
	})

	b.Run("ContextMemoryFootprint", func(b *testing.B) {
		contexts := make([]*domain.ProcessingContext, b.N)
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			contexts[i] = domain.NewProcessingContext(fmt.Sprintf("mem-ctx-%d", i))
			
			// Add multiple messages to simulate conversation
			for j := 0; j < 10; j++ {
				contexts[i].AddUserMessage(fmt.Sprintf("User message %d in context %d", j, i))
				contexts[i].AddAssistantMessage(fmt.Sprintf("Assistant response %d in context %d", j, i))
			}
			
			// Add metadata
			for j := 0; j < 5; j++ {
				contexts[i].SetMetadata(fmt.Sprintf("key%d", j), fmt.Sprintf("value%d", j))
			}
		}
		
		_ = contexts
	})
}

// BenchmarkLargePayloads tests performance with large data
func BenchmarkLargePayloads(b *testing.B) {
	repo, cleanup := setupBenchmarkDatabase(b)
	defer cleanup()

	ctx := context.Background()

	// Test different payload sizes
	sizes := []int{1024, 10240, 102400, 1048576} // 1KB, 10KB, 100KB, 1MB
	
	for _, size := range sizes {
		b.Run(fmt.Sprintf("Payload%dBytes", size), func(b *testing.B) {
			largeData := string(make([]byte, size))
			
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				message := domain.NewMessage(
					fmt.Sprintf("large-msg-%d-%d", size, i),
					domain.MessageTypeWorkRequest,
					map[string]interface{}{
						"code": largeData,
						"task": "analyze large payload",
					},
				)
				
				err := repo.SaveMessage(ctx, message)
				if err != nil {
					b.Fatal(err)
				}
				
				// Also test retrieval
				retrieved, err := repo.GetMessageByID(ctx, message.ID)
				if err != nil {
					b.Fatal(err)
				}
				_ = retrieved
			}
		})
	}
}

// BenchmarkCleanupOperations tests cleanup performance
func BenchmarkCleanupOperations(b *testing.B) {
	repo, cleanup := setupBenchmarkDatabase(b)
	defer cleanup()

	logConfig := &infrastructure.LoggingConfig{
		Level:  "error",
		Format: "text",
		Output: "stdout",
	}
	logger := infrastructure.NewLogger(logConfig)

	claudeService := &BenchmarkClaudeService{}
	requestService := usecase.NewRequestService(repo, claudeService, logger)
	messageProcessor := usecase.NewMessageProcessor(repo, repo, requestService, logger)

	ctx := context.Background()

	// Create expired contexts for cleanup
	for i := 0; i < 100; i++ {
		context := domain.NewProcessingContext(fmt.Sprintf("cleanup-ctx-%d", i))
		context.LastUsedAt = time.Now().Add(-2 * time.Hour) // Make it expired
		err := repo.SaveProcessingContext(ctx, context)
		require.NoError(b, err)
	}

	b.Run("ExpiredContextCleanup", func(b *testing.B) {
		maxAge := 1 * time.Hour
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			err := messageProcessor.CleanupExpiredContexts(ctx, maxAge)
			if err != nil {
				b.Fatal(err)
			}
		}
	})
}

// Performance benchmarks provide the following insights:
// - Target: Message processing < 100ms (excluding Claude API call)
// - Target: Database operations < 10ms
// - Target: Memory usage < 50MB for normal operations
// - Target: Concurrent processing should handle 100+ requests/second

// Example benchmark results to aim for:
// BenchmarkMessageCreation-8            	 2000000	       800 ns/op
// BenchmarkDatabaseOperations/MessageSave-8    	   10000	    150000 ns/op
// BenchmarkMessageProcessing/SingleMessage-8    	      20	  55000000 ns/op (includes 50ms Claude simulation)
// BenchmarkContextOperations/ContextSave-8      	    1000	   1800000 ns/op