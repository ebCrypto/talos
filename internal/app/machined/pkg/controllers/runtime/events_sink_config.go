// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at http://mozilla.org/MPL/2.0/.

package runtime

import (
	"context"
	"fmt"

	"github.com/cosi-project/runtime/pkg/controller"
	"github.com/cosi-project/runtime/pkg/safe"
	"github.com/cosi-project/runtime/pkg/state"
	"github.com/siderolabs/go-procfs/procfs"
	"go.uber.org/zap"

	"github.com/siderolabs/talos/pkg/machinery/constants"
	"github.com/siderolabs/talos/pkg/machinery/resources/runtime"
)

// EventsSinkConfigController generates configuration for kmsg log delivery.
type EventsSinkConfigController struct {
	Cmdline *procfs.Cmdline
}

// Name implements controller.Controller interface.
func (ctrl *EventsSinkConfigController) Name() string {
	return "runtime.EventsSinkConfigController"
}

// Inputs implements controller.Controller interface.
func (ctrl *EventsSinkConfigController) Inputs() []controller.Input {
	return nil
}

// Outputs implements controller.Controller interface.
func (ctrl *EventsSinkConfigController) Outputs() []controller.Output {
	return []controller.Output{
		{
			Type: runtime.EventSinkConfigType,
			Kind: controller.OutputExclusive,
		},
	}
}

// Run implements controller.Controller interface.
//
//nolint:gocyclo,cyclop
func (ctrl *EventsSinkConfigController) Run(ctx context.Context, r controller.Runtime, logger *zap.Logger) (err error) {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-r.EventCh():
		}

		var endpoint string

		if ctrl.Cmdline != nil {
			if val := ctrl.Cmdline.Get(constants.KernelParamEventsSink).First(); val != nil {
				endpoint = *val
			}
		}

		if endpoint == "" {
			if err := r.Destroy(ctx, runtime.NewEventSinkConfig().Metadata()); err != nil && !state.IsNotFoundError(err) {
				return fmt.Errorf("error destroying event sink config: %w", err)
			}
		} else {
			if err := safe.WriterModify(ctx, r, runtime.NewEventSinkConfig(), func(cfg *runtime.EventSinkConfig) error {
				cfg.TypedSpec().Endpoint = endpoint

				return nil
			}); err != nil {
				return fmt.Errorf("error updating kmsg log config: %w", err)
			}
		}

		r.ResetRestartBackoff()
	}
}
