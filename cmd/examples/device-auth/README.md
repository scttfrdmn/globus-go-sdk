<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- Copyright (c) 2025 Scott Friedman and Project Contributors -->

# Device Authentication Flow Example

This example demonstrates how to implement device authentication flow using the Globus SDK. Device authentication flow is designed for command-line tools, scripts, and other non-web applications that cannot use a web browser for authentication.

## Overview

The device authentication flow consists of the following steps:

1. The application requests a device code from Globus Auth
2. Globus Auth returns a device code, user code, and verification URL
3. The user is prompted to visit the verification URL and enter the user code
4. The application polls Globus Auth to check if the user has completed the authorization
5. Once the user authorizes the application, Globus Auth returns the access token and refresh token

## Running the Example

To run this example, you need to set the `GLOBUS_CLIENT_ID` environment variable:

```bash
export GLOBUS_CLIENT_ID=your-client-id
go run main.go
```

## Implementation Details

The device authentication flow is implemented using the following methods from the `auth.Client`:

- `RequestDeviceCode`: Initiates the device flow and returns a `DeviceCodeResponse`
- `PollDeviceCode`: Polls Globus Auth to check if the user has completed authorization
- `CompleteDeviceFlow`: A convenience method that handles the entire flow

The `CompleteDeviceFlow` method is the easiest way to implement device flow as it handles all the steps automatically, including polling with appropriate intervals and error handling.

## Example Output

```
Starting device authentication flow...

===== Device Authorization Required =====
Please visit this URL to authorize this application:
  https://auth.globus.org/device
  
Enter the following code when prompted:
  ABCD-1234
=======================================
This code will expire in 300 seconds.
Waiting for authorization...

Authentication successful!
Access Token: eyJhbGciOiJSUzI...
Token expires in: 172800 seconds
Refresh Token: AQEAAAAAAACXUK...
```

## Error Handling

The device flow can result in several specific error types:

- `authorization_pending`: The user has not yet completed the authorization (normal during polling)
- `slow_down`: The application is polling too frequently and should reduce the polling rate
- `expired_token`: The device code has expired; a new one should be requested
- `access_denied`: The user denied the authorization request

These errors are represented as `DeviceAuthError` objects and can be checked using the `IsDeviceAuthError` function or the more specific helpers like `IsAuthorizationPending`, `IsSlowDown`, etc.