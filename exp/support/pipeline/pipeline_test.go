package pipeline_test

import (
	"context"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/diamnet/go/exp/support/pipeline"
	"github.com/diamnet/go/support/errors"
	"github.com/stretchr/testify/assert"
)

func TestPipelineCanBeProcessedAgain(t *testing.T) {
	p := pipeline.New(
		pipeline.Node(&NoOpProcessor{}),
	)

	assert.NoError(t, <-p.Process(&SimpleReader{CountObject: 10}))
	assert.NoError(t, <-p.Process(&SimpleReader{CountObject: 20}))
}

func TestCannotRunProcessOnRunningPipeline(t *testing.T) {
	p := pipeline.New(
		pipeline.Node(&NoOpProcessor{}),
	)

	go p.Process(&SimpleReader{})
	defer p.Shutdown()
	time.Sleep(100 * time.Millisecond)
	assert.Panics(t, func() {
		p.Process(&SimpleReader{})
	})
}

func TestNoErrorsOnSuccess(t *testing.T) {
	p := pipeline.New(
		pipeline.Node(&NoOpProcessor{}),
	)

	assert.NoError(t, <-p.Process(&SimpleReader{CountObject: 10}))
}

func TestErrorsOnFailure(t *testing.T) {
	p := pipeline.New(
		pipeline.Node(&NoOpProcessor{ReturnError: true}),
	)

	err := <-p.Process(&SimpleReader{CountObject: 10})
	assert.Error(t, err)
	assert.Equal(t, "Processor NoOpProcessor errored: Test error", err.Error())
}

func TestHooksCalled(t *testing.T) {
	p := pipeline.New(
		pipeline.Node(&NoOpProcessor{}),
	)

	preHookCalled := false
	p.AddPreProcessingHook(func(ctx context.Context) (context.Context, error) {
		preHookCalled = true
		return ctx, nil
	})

	postHookCalled := false
	p.AddPostProcessingHook(func(ctx context.Context, err error) error {
		postHookCalled = true
		return nil
	})

	err := <-p.Process(&SimpleReader{CountObject: 10})
	assert.NoError(t, err)
	assert.True(t, preHookCalled, "pre-hook not called")
	assert.True(t, postHookCalled, "post-hook not called")
}

func TestPostHooksCalledWithError(t *testing.T) {
	p := pipeline.New(
		pipeline.Node(&NoOpProcessor{ReturnError: true}),
	)

	errChan := make(chan error, 1)

	p.AddPostProcessingHook(func(ctx context.Context, err error) error {
		errChan <- err
		return nil
	})

	err := <-p.Process(&SimpleReader{CountObject: 10})
	assert.Error(t, err)
	assert.Equal(t, "Processor NoOpProcessor errored: Test error", err.Error())

	hookErr := <-errChan
	assert.Error(t, hookErr)
	assert.Equal(t, "Processor NoOpProcessor errored: Test error", hookErr.Error())
}

func TestProcessReturnsErrorWhenPostHooksErrors(t *testing.T) {
	p := pipeline.New(
		pipeline.Node(&NoOpProcessor{}),
	)

	p.AddPostProcessingHook(func(ctx context.Context, err error) error {
		return errors.New("post-hook error")
	})

	err := <-p.Process(&SimpleReader{CountObject: 10})
	assert.Error(t, err)
	assert.Equal(t, "Error running post-hook: post-hook error", err.Error())
}

func TestPostHookNotCalledWhenShutDown(t *testing.T) {
	p := pipeline.New(
		pipeline.Node(&NoOpProcessor{}),
	)
	p.AddPostProcessingHook(func(ctx context.Context, err error) error {
		panic("Hook shouldn't be called!")
	})

	go p.Process(&SimpleReader{})
	time.Sleep(100 * time.Millisecond)
	p.Shutdown()
}

// SimpleReader sends CountObject objects. If CountObject = 0 it
// streams infinite number of objects.
type SimpleReader struct {
	sync.Mutex
	CountObject int

	sent int
}

func (r *SimpleReader) GetContext() context.Context {
	return context.Background()
}

func (r *SimpleReader) Read() (interface{}, error) {
	r.Lock()
	defer r.Unlock()

	if r.CountObject != 0 && r.sent == r.CountObject {
		return nil, io.EOF
	}

	r.sent++
	return "test", nil
}

func (r *SimpleReader) Close() error {
	return nil
}

type NoOpProcessor struct {
	ReturnError bool
}

func (p *NoOpProcessor) Process(ctx context.Context, store *pipeline.Store, r pipeline.Reader, w pipeline.Writer) error {
	defer r.Close()
	defer w.Close()

	for {
		_, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return errors.Wrap(err, "Error reading from Reader")
			}
		}

		if p.ReturnError {
			return errors.New("Test error")
		}

		select {
		case <-ctx.Done():
			return nil
		default:
			continue
		}
	}

	return nil
}

func (p *NoOpProcessor) Name() string {
	return "NoOpProcessor"
}

func (p *NoOpProcessor) Reset() {}
