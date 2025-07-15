package prompt

import (
	"context"
	"errors"
	"testing"
	"time"
)

// mockContent 用于模拟 ResourceContents 实现
type mockContent struct {
	Data string
}

func (mockContent) isResourceContents() {}

// mockHandler 返回一个简单的 mockContent
func mockHandler(ctx context.Context, req *ReadResourceRequest) (ResourceContents, error) {
	if req.Params.URI == "invalid" {
		return nil, errors.New("resource not found")
	}
	return mockContent{Data: "mock data"}, nil
}

func TestResourceManager_RegisterAndGet(t *testing.T) {
	manager := newResourceManager()
	resource := &Resource{
		Name: "test",
		URI:  "res://test",
	}

	manager.registerResource(resource, mockHandler)
	got, exists := manager.getResource("res://test")
	if !exists {
		t.Fatal("expected resource to exist")
	}
	if got.Name != "test" {
		t.Errorf("expected name 'test', got '%s'", got.Name)
	}
}

func TestResourceManager_ReadResource(t *testing.T) {
	manager := newResourceManager()
	resource := &Resource{Name: "test", URI: "res://test"}
	manager.registerResource(resource, mockHandler)

	req := &JSONRPCRequest{
		ID:      "1",
		JSONRPC: "2.0",
		Params: map[string]interface{}{
			"uri": "res://test",
		},
	}

	resp, err := manager.handleReadResource(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	result, ok := resp.(ReadResourceResult)
	if !ok {
		t.Fatal("expected ReadResourceResult type")
	}
	if len(result.Contents) != 1 {
		t.Errorf("expected 1 content, got %d", len(result.Contents))
	}
}

func TestResourceManager_SubscribeUnsubscribe(t *testing.T) {
	manager := newResourceManager()
	uri := "res://test"
	ch := manager.subscribe(uri)

	if len(manager.subscribers[uri]) != 1 {
		t.Fatal("subscriber not added")
	}

	manager.unsubscribe(uri, ch)

	if _, ok := manager.subscribers[uri]; ok {
		t.Fatal("subscriber not removed")
	}
}

func TestResourceManager_NotifyUpdate(t *testing.T) {
	manager := newResourceManager()
	uri := "res://update"
	ch := manager.subscribe(uri)

	// 使用 goroutine 监听通知
	done := make(chan bool)
	go func() {
		select {
		case n := <-ch:
			if n.Method != "notifications/resources/updated" {
				t.Errorf("unexpected method: %s", n.Method)
			}
			done <- true
		case <-time.After(1 * time.Second):
			t.Error("notification not received")
			done <- false
		}
	}()

	manager.notifyUpdate(uri)
	<-done
}

func TestResourceManager_ListResources(t *testing.T) {
	manager := newResourceManager()
	manager.registerResource(&Resource{Name: "a", URI: "res://a"}, mockHandler)
	manager.registerResource(&Resource{Name: "b", URI: "res://b"}, mockHandler)

	req := &JSONRPCRequest{
		ID:      "list",
		JSONRPC: "2.0",
	}

	resp, err := manager.handleListResources(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	result, ok := resp.(ListResourcesResult)
	if !ok {
		t.Fatal("unexpected result type")
	}
	if len(result.Resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(result.Resources))
	}
}
