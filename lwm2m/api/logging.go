// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

// +build !test

package api

import (
	"context"
	"fmt"
	"time"

	"github.com/mainflux/fluxm2m/lwm2m"
	log "github.com/mainflux/mainflux/logger"
)

var _ lwm2m.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger log.Logger
	svc    lwm2m.Service
}

// LoggingMiddleware adds logging facilities to the adapter.
func LoggingMiddleware(svc lwm2m.Service, logger log.Logger) lwm2m.Service {
	return &loggingMiddleware{logger, svc}
}

func (lm *loggingMiddleware) Publish(ctx context.Context) (err error) {
	defer func(begin time.Time) {
		if err != nil {
			lm.logger.Warn(fmt.Sprintf("with error: %s.", err))
			return
		}
		lm.logger.Info("%s without errors.")
	}(time.Now())

	return lm.svc.Publish(ctx)
}
