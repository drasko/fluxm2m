// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"fmt"

	log "github.com/mainflux/mainflux/logger"

	lwm2m "github.com/drasko/fluxm2m"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
)

var (
	logger  log.Logger
	service lwm2m.Service
)

// MakeCoAPHandler creates handler for CoAP messages.
func MakeCoAPHandler(svc lwm2m.Service, l log.Logger) mux.HandlerFunc {
	logger = l
	service = svc

	return handler
}

func sendResp(w mux.ResponseWriter, resp *message.Message) {
	if err := w.Client().WriteMessage(resp); err != nil {
		logger.Warn(fmt.Sprintf("Can't set response: %s", err))
	}
}

func handler(w mux.ResponseWriter, m *mux.Message) {
	resp := message.Message{
		Code:    codes.Content,
		Token:   m.Token,
		Context: m.Context,
		Options: make(message.Options, 0, 16),
	}
	defer sendResp(w, &resp)

	var err error

	if m.Options == nil {
		logger.Warn("Nil options")
		resp.Code = codes.BadOption
		return
	}

	switch m.Code {
	case codes.GET:
		var obs uint32
		obs, err = m.Options.Observe()
		if err != nil {
			resp.Code = codes.BadOption
			logger.Warn(fmt.Sprintf("Error reading observe option: %s", err))
			return
		}
		if obs == 0 {
			err = service.Subscribe(context.Background())
			break
		}
		service.Unsubscribe(context.Background(), m.Token.String())
	case codes.POST:
		err = service.Publish(context.Background())
	default:
		resp.Code = codes.NotFound
		return
	}
	if err != nil {
		switch {
		case errors.Contains(err, lwm2m.ErrUnauthorized):
			resp.Code = codes.Unauthorized
			return
		case errors.Contains(err, lwm2m.ErrUnsubscribe):
			resp.Code = codes.InternalServerError
		}
	}
}
