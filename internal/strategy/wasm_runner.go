package strategy

import (
	"context"
	"fmt"
	"market-ingestor/internal/model"

	"github.com/tetratelabs/wazero"
	"github.com/tetratelabs/wazero/api"
)

type WasmRunner struct {
	runtime wazero.Runtime
	mod     api.Module
}

func NewWasmRunner(ctx context.Context, wasmCode []byte) (*WasmRunner, error) {
	r := wazero.NewRuntime(ctx)

	// Compile and instantiate the module
	mod, err := r.Instantiate(ctx, wasmCode)
	if err != nil {
		r.Close(ctx)
		return nil, fmt.Errorf("failed to instantiate wasm: %w", err)
	}

	return &WasmRunner{
		runtime: r,
		mod:     mod,
	}, nil
}

func (r *WasmRunner) OnCandle(ctx context.Context, candle model.KLine) (Action, error) {
	// 1. Get the handle to the OnCandle function in Wasm
	onCandle := r.mod.ExportedFunction("OnCandle")
	if onCandle == nil {
		return ActionHold, fmt.Errorf("wasm module does not export OnCandle")
	}

	// 2. Map Go candle to Wasm memory (simplified example)
	price, _ := candle.Close.Float64()

	// 3. Call the function
	results, err := onCandle.Call(ctx, api.EncodeF64(price))
	if err != nil {
		return ActionHold, fmt.Errorf("failed to call OnCandle: %w", err)
	}

	// 4. Map the return value back to Action
	if len(results) > 0 {
		// Define a mapping between Wasm integer returns and our string Actions
		switch results[0] {
		case 1:
			return ActionBuy, nil
		case 2:
			return ActionSell, nil
		default:
			return ActionHold, nil
		}
	}

	return ActionHold, nil
}

func (r *WasmRunner) Close(ctx context.Context) error {
	return r.runtime.Close(ctx)
}
