package callbacks

import (
	"context"
	"errors"
	"sync"

	"github.com/cloudwego/eino/adk"
	einocallbacks "github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components"
	einomodel "github.com/cloudwego/eino/components/model"
	einotool "github.com/cloudwego/eino/components/tool"
)

// FakeTestRecorder is a test fixture for proving the evaluator callback seam.
//
// It is intentionally execution-local and must not become a production
// observability abstraction.
//
// The recorder uses a mutex because callback handlers run during live Eino
// execution, where model and tool lifecycle callbacks should be treated as
// concurrent with test assertions and with one another. Even though current
// fixture executions are small and deterministic, the callback seam itself is
// asynchronous enough that recording without synchronization would make the
// test fixture racy and would misrepresent the production callback boundary.
type FakeTestRecorder struct {
	mu sync.Mutex

	constructed int
	attached    int
	agentStarts int
	agentEnds   int
	modelStarts int
	modelEnds   int
	toolStarts  int
	toolEnds    int
}

// FakeTestSnapshot is the immutable view of the recorded callback facts.
type FakeTestSnapshot struct {
	Constructed int
	Attached    int
	AgentStarts int
	AgentEnds   int
	ModelStarts int
	ModelEnds   int
	ToolStarts  int
	ToolEnds    int
}

// Snapshot returns the recorded callback facts for assertions.
//
// The snapshot is taken under the same lock used by the callback handlers so
// tests read a consistent view instead of racing against in-flight callback
// updates.
func (r *FakeTestRecorder) Snapshot() FakeTestSnapshot {
	if r == nil {
		return FakeTestSnapshot{}
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	return FakeTestSnapshot{
		Constructed: r.constructed,
		Attached:    r.attached,
		AgentStarts: r.agentStarts,
		AgentEnds:   r.agentEnds,
		ModelStarts: r.modelStarts,
		ModelEnds:   r.modelEnds,
		ToolStarts:  r.toolStarts,
		ToolEnds:    r.toolEnds,
	}
}

// NewFakeTestCallbackFactory returns a cold callback factory for evaluator
// tests.
//
// The factory increments Constructed under lock to keep setup accounting
// consistent with later execution-time callback writes. The returned handler
// also updates the recorder under lock because Eino may emit agent, model, and
// tool callback events while the test is concurrently waiting on or asserting
// over recorder state.
func NewFakeTestCallbackFactory(recorder *FakeTestRecorder) Factory {
	return func(context.Context) (einocallbacks.Handler, error) {
		if recorder == nil {
			return nil, errors.New("fake test callback recorder is required")
		}

		recorder.mu.Lock()
		recorder.constructed++
		recorder.mu.Unlock()

		return einocallbacks.NewHandlerBuilder().
			OnStartFn(func(ctx context.Context, info *einocallbacks.RunInfo, input einocallbacks.CallbackInput) context.Context {
				if info == nil {
					return ctx
				}

				recorder.mu.Lock()
				defer recorder.mu.Unlock()

				switch info.Component {
				case adk.ComponentOfAgent:
					recorder.attached++
					recorder.agentStarts++
				case components.ComponentOfChatModel:
					if einomodel.ConvCallbackInput(input) != nil {
						recorder.modelStarts++
					}
				case components.ComponentOfTool:
					if einotool.ConvCallbackInput(input) != nil {
						recorder.toolStarts++
					}
				}

				return ctx
			}).
			OnEndFn(func(ctx context.Context, info *einocallbacks.RunInfo, output einocallbacks.CallbackOutput) context.Context {
				if info == nil {
					return ctx
				}

				recorder.mu.Lock()
				defer recorder.mu.Unlock()

				switch info.Component {
				case adk.ComponentOfAgent:
					recorder.agentEnds++
				case components.ComponentOfChatModel:
					if einomodel.ConvCallbackOutput(output) != nil {
						recorder.modelEnds++
					}
				case components.ComponentOfTool:
					if einotool.ConvCallbackOutput(output) != nil {
						recorder.toolEnds++
					}
				}

				return ctx
			}).
			Build(), nil
	}
}
