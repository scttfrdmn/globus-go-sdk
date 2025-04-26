# Globus Go SDK Web Application Example

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

This example demonstrates using the Globus Go SDK in a web application, showcasing authentication flows, token management, and integration with Globus services like Flows and Search.

## Features

- OAuth2 authentication with Globus
- User sessions management
- Token storage with automatic refreshing
- Integration with Globus Flows
- Integration with Globus Search
- Simple web interface

## Prerequisites

To run this example, you need:

1. Go 1.16 or later
2. A Globus developer account
3. A registered Globus app with:
   - Client ID and Client Secret
   - Redirect URL: `http://localhost:8080/callback`
   - Required scopes:
     - `openid`
     - `profile`
     - `email`
     - `https://auth.globus.org/scopes/eec9b274-0c81-4334-bdc2-54e90e689b9a/manage_flows`
     - `https://auth.globus.org/scopes/search.api.globus.org/all`

## Setup

1. Clone the repository:

```bash
git clone https://github.com/yourusername/globus-go-sdk.git
cd globus-go-sdk
```

2. Set up environment variables:

```bash
export GLOBUS_CLIENT_ID="your-client-id"
export GLOBUS_CLIENT_SECRET="your-client-secret"
export SESSION_SECRET="random-secret-for-sessions"
```

3. Run the application:

```bash
cd examples/webapp
go run main.go
```

4. Open your browser and navigate to `http://localhost:8080`

## How it Works

### Authentication Flow

1. User clicks "Login with Globus"
2. User is redirected to Globus Auth for authentication
3. After successful authentication, Globus redirects back with an authorization code
4. The app exchanges the code for tokens using the SDK
5. Tokens are stored securely for future API calls

### Token Management

- Tokens are stored in the filesystem (in the `tokens` directory)
- The token manager automatically refreshes expired tokens
- Access tokens are obtained for specific scopes when needed

### Service Integration

- **Flows Service**: The app displays a list of the user's flows
- **Search Service**: Users can search for data using Globus Search

## Architecture

The application uses a simple MVC architecture:

- **Model**: Session and token management
- **View**: HTML templates rendered server-side
- **Controller**: HTTP handlers for different routes

## Security Considerations

This example implements several security measures:

- CSRF protection for the OAuth flow using state parameters
- HttpOnly cookies for session management
- Secure token storage
- Access token scoping

However, for production use, additional security measures should be implemented:

- HTTPS for all communications
- More secure session management
- Rate limiting
- Input validation
- Proper error handling and logging

## Next Steps

To extend this example:

1. Add more Globus services like Transfer
2. Implement more robust error handling
3. Add logging and monitoring
4. Enhance the UI with more interactive features
5. Implement proper user sessions with database storage

## License

This example is licensed under the Apache License 2.0. See the LICENSE file for details.