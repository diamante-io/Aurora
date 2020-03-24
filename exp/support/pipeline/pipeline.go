package pipeline

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hcnet/go/support/errors"
)

func New(rootProcessor *PipelineNode) *Pipeline {
	return &Pipeline{root: rootProcessor}
}

func Node(processor Processor) *PipelineNode {
	return &PipelineNode{
		Processor: processor,
	}
}

func (p *Pipeline) PrintStatus() {
	p.printNodeStatus(p.root, 0)
}

func (p *Pipeline) AddPreProcessingHook(hook func(context.Context) (context.Context, error)) {
	p.preProcessingHooks = append(p.preProcessingHooks, hook)
}

func (p *Pipeline) AddPostProcessingHook(hook func(context.Context, error) error) {
	p.postProcessingHooks = append(p.postProcessingHooks, hook)
}

func (p *Pipeline) printNodeStatus(node *PipelineNode, level int) {
	fmt.Print(strings.Repeat("  ", level))

	var wrRatio = float32(0)
	if node.readEntries > 0 {
		wrRatio = float32(node.wroteEntries) / float32(node.readEntries)
	}

	icon := ""
	if node.queuedEntries > bufferSize/10*9 {
		icon = "⚠️ "
	}

	fmt.Printf(
		"└ %s%s read=%d (queued=%d rps=%d) wrote=%d (w/r ratio = %1.5f)\n",
		icon,
		node.Processor.Name(),
		node.readEntries,
		node.queuedEntries,
		node.readsPerSecond,
		node.wroteEntries,
		wrRatio,
	)

	for _, child := range node.Children {
		p.printNodeStatus(child, level+1)
	}
}

func (p *Pipeline) SetRoot(rootProcessor *PipelineNode) {
	p.root = rootProcessor
}

// setRunning protects from processing more than once at a time.
func (p *Pipeline) setRunning(setRunning bool) {
	if setRunning {
		if p.running {
			panic("Pipeline is running...")
		}

		if p.shutDown {
			panic("Pipeline was shut down...")
		}
	}

	p.running = setRunning
}

// reset resets internal state of the pipeline and all the nodes and processors.
func (p *Pipeline) reset() {
	p.cancelled = false
	p.resetNode(p.root)
}

func (p *Pipeline) sendPreProcessingHooks(ctx context.Context) (context.Context, error) {
	var err error

	for _, hook := range p.preProcessingHooks {
		ctx, err = hook(ctx)
		if err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

func (p *Pipeline) sendPostProcessingHooks(ctx context.Context, processingError error) error {
	for _, hook := range p.postProcessingHooks {
		err := hook(ctx, processingError)
		if err != nil {
			return err
		}
	}

	return nil
}

// resetNode resets internal state of the pipeline node and internal processor and
// calls itself recursively on all of the children.
func (p *Pipeline) resetNode(node *PipelineNode) {
	node.reset()
	for _, child := range node.Children {
		p.resetNode(child)
	}
}

func (p *Pipeline) Process(reader Reader) <-chan error {
	// Protects internal fields
	p.mutex.Lock()
	defer p.mutex.Unlock()

	p.setRunning(true)
	p.reset()

	ctx, err := p.sendPreProcessingHooks(reader.GetContext())
	if err != nil {
		p.setRunning(false)
		errorChan := make(chan error, 1)
		errorChan <- errors.Wrap(err, "Error running pre-hook")
		return errorChan
	}

	ctx, p.cancelFunc = context.WithCancel(ctx)
	return p.processStateNode(ctx, &Store{}, p.root, reader)
}

func (p *Pipeline) processStateNode(ctx context.Context, store *Store, node *PipelineNode, reader Reader) <-chan error {
	outputs := make([]Writer, len(node.Children))

	for i := range outputs {
		outputs[i] = &BufferedReadWriter{
			context: reader.GetContext(),
		}
	}

	var wg sync.WaitGroup

	writer := &multiWriter{
		writers:    outputs,
		closeAfter: 1,
	}

	var processingError error

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := node.Processor.Process(ctx, store, reader, writer)
		if err != nil {
			// Protects from cancelling twice and sending multiple errors to err channel
			p.mutex.Lock()
			defer p.mutex.Unlock()

			if p.cancelled {
				return
			}

			wrappedErr := errors.Wrap(err, fmt.Sprintf("Processor %s errored", node.Processor.Name()))

			p.cancelled = true
			p.cancelFunc()
			processingError = wrappedErr
		}
	}()

	finishUpdatingStats := p.updateStats(node, reader, writer)

	for i, child := range node.Children {
		wg.Add(1)
		go func(i int, child *PipelineNode) {
			defer wg.Done()
			err := <-p.processStateNode(ctx, store, child, outputs[i].(*BufferedReadWriter))
			if err != nil {
				processingError = err
			}
		}(i, child)
	}

	errorChan := make(chan error, 1)

	go func() {
		wg.Wait()
		finishUpdatingStats <- true

		if node == p.root {
			// If pipeline processing is finished run post-hooks (if not shut down)
			// and send error if not already sent.
			var returnError error
			if p.shutDown {
				returnError = nil
			} else {
				err := p.sendPostProcessingHooks(reader.GetContext(), processingError)
				if err != nil {
					returnError = errors.Wrap(err, "Error running post-hook")
				} else {
					returnError = processingError
				}
			}

			p.mutex.Lock()
			p.setRunning(false)
			p.mutex.Unlock()

			errorChan <- returnError
		} else {
			// For non-root node just send an error
			errorChan <- processingError
		}
	}()

	return errorChan
}

// Shutdown stops the processing. Please note that post-processing hooks will not
// be executed when Shutdown() is called.
func (p *Pipeline) Shutdown() {
	// Protects internal fields
	p.mutex.Lock()
	defer p.mutex.Unlock()

	if p.cancelled {
		return
	}
	p.shutDown = true
	p.cancelled = true
	p.cancelFunc()
}

func (p *Pipeline) updateStats(node *PipelineNode, reader Reader, writer *multiWriter) chan<- bool {
	// Update stats
	interval := time.Second
	done := make(chan bool)
	ticker := time.NewTicker(interval)

	go func() {
		defer ticker.Stop()

		for {
			// This is not thread-safe: check if Mutex slows it down a lot...
			readBuffer, readBufferIsBufferedReadWriter := reader.(*BufferedReadWriter)

			node.writesPerSecond = (writer.wroteEntries - node.wroteEntries) * int(time.Second/interval)
			node.wroteEntries = writer.wroteEntries

			if readBufferIsBufferedReadWriter {
				node.readsPerSecond = (readBuffer.readEntries - node.readEntries) * int(time.Second/interval)
				node.readEntries = readBuffer.readEntries
				node.queuedEntries = readBuffer.QueuedEntries()
			}

			select {
			case <-ticker.C:
				continue
			case <-done:
				// Pipeline done
				return
			}
		}
	}()

	return done
}
