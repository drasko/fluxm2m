// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package lwm2m

import (
	"context"
	"errors"
)

// Exported errors
var (
	ErrUnauthorized = errors.New("unauthorized access")
	ErrUnsubscribe  = errors.New("unable to unsubscribe")
)

// Service specifies lwm2m service API.
type Service interface {
	// Subscribe
	Subscribe(ctx context.Context) error

	// Unsubscribe
	Unsubscribe(ctx context.Context) error

	// Publish
	Publish(ctx context.Context) error
}

var _ Service = (*lwm2mService)(nil)

// Observers is a map of maps,
type lwm2mService struct {
}

// New instantiates the CoAP adapter implementation.
func New() Service {
	s := &lwm2mService{}

	return s
}

func (svc *lwm2mService) Subscribe(ctx context.Context) error {
	return nil
}

func (svc *lwm2mService) Unsubscribe(ctx context.Context, token string) error {
	return nil
}

func (svc *lwm2mService) Publish(ctx context.Context) error {
	return nil
}
