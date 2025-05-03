// SPDX-License-Identifier: Apache-2.0
// Copyright (c) 2025 Scott Friedman and Project Contributors
package core

import (
	"log"
	"os"

	"github.com/scttfrdmn/globus-go-sdk/pkg/core/interfaces"
	"github.com/scttfrdmn/globus-go-sdk/pkg/core/transport"
)

// InitTransport initializes a transport for a client
// This function helps break the import cycle between core and transport packages
func InitTransport(client interfaces.ClientInterface, debug, trace bool) interfaces.Transport {
	loggerForTransport := log.New(os.Stderr, "", log.LstdFlags)
	
	// Create deferred transport first
	dt := transport.NewDeferredTransport(&transport.Options{
		Debug:  debug,
		Trace:  trace,
		Logger: loggerForTransport,
	})
	
	// Now attach the client to create the actual transport
	return dt.AttachClient(client)
}