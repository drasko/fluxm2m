// Copyright (c) Mainflux
// SPDX-License-Identifier: Apache-2.0

package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-zoo/bone"
	"github.com/mainflux/mainflux"
	"github.com/mainflux/mainflux/errors"
	log "github.com/mainflux/mainflux/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/mainflux/fluxm2m/lwm2m"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
)

const (
	protocol = "lwm2m"
)

var (
	logger  log.Logger
	service lwm2m.Service
)

//MakeHTTPHandler creates handler for version endpoint.
func MakeHTTPHandler() http.Handler {
	b := bone.New()
	b.GetFunc("/version", mainflux.Version(protocol))
	b.Handle("/metrics", promhttp.Handler())

	return b
}

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
		_, err = m.Options.Observe()
		if err != nil {
			resp.Code = codes.BadOption
			logger.Warn(fmt.Sprintf("Error reading observe option: %s", err))
			return
		}
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
