package mock

import (
	"context"
	"net/http"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestBenchmarkServer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	readyChan := make(chan struct{})
	errChan := make(chan error)

	var wg sync.WaitGroup
	wg.Add(1)
	NewBenchmarkServer(ctx, &wg, readyChan, errChan)

	select {
	case <-readyChan:
	case err := <-errChan:
		t.Fatalf("Could not start benchmark server: %v", err)
	}

	// We need to wait some time for nginx to acutally start running
	time.Sleep(5 * time.Second)
	resp, err := http.Get("http://127.0.0.1:8080")
	require.Nil(t, err)

	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	cancel()

	wg.Wait()
}
