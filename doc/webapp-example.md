# Web Application Example

<!-- SPDX-License-Identifier: Apache-2.0 -->
<!-- SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors -->

This guide explains the web application example that demonstrates how to use the Globus Go SDK in a web-based environment.

## Overview

The web application example showcases several key features:

1. Implementing OAuth 2.0 flows with Globus Auth
2. Managing user sessions
3. Storing and refreshing tokens
4. Integrating with multiple Globus services (Flows and Search)
5. Building a simple but functional web UI

## Key Components

### Authentication Flow

The example implements the Authorization Code flow:

1. User clicks "Log in with Globus"
2. Application redirects to Globus Auth with the necessary parameters:
   - Client ID
   - Redirect URI
   - Requested scopes
   - State parameter for CSRF protection
3. User authenticates on Globus Auth
4. Globus Auth redirects back to the application with an authorization code
5. Application exchanges the code for tokens
6. Application creates a user session and stores tokens

### Session Management

The example implements a simple in-memory session store:

```go
// Session represents a user session
type Session struct {
    ID           string    `json:"id"`
    UserID       string    `json:"user_id"`
    Username     string    `json:"username"`
    Email        string    `json:"email"`
    SessionStart time.Time `json:"session_start"`
    LastAccess   time.Time `json:"last_access"`
}

// Application state
type App struct {
    // ...
    Sessions      map[string]*Session
    // ...
}
```

A session cookie is used to identify the user's session, and the cookie is set to HttpOnly to enhance security.

### Token Management

The example uses the SDK's token storage and management capabilities:

```go
// Initialize token storage
app.TokenStorage, err = tokens.NewFileStorage(tokensDir)
if err != nil {
    log.Fatalf("Failed to initialize token storage: %v", err)
}

// Initialize auth client
app.AuthClient = auth.NewClient(config.ClientID, config.ClientSecret)
app.AuthClient.SetRedirectURL(config.RedirectURL)

// Initialize token manager
app.TokenManager = tokens.NewManager(app.TokenStorage, app.AuthClient)
```

The application stores tokens securely on the filesystem and refreshes them automatically when needed.

### Service Integration

The example integrates with two Globus services:

**Flows Client:**
```go
// Get tokens for the user
accessToken, err := app.getAccessToken(ctx, session.UserID, flows.FlowsScope)
if err != nil {
    // Handle error
}

// Initialize Flows client
app.FlowsClient = flows.NewClient(accessToken)

// List flows
flowList, err := app.FlowsClient.ListFlows(ctx, &flows.ListFlowsOptions{
    Limit: 10,
})
```

**Search Client:**
```go
// Get tokens for the user
accessToken, err := app.getAccessToken(ctx, session.UserID, search.SearchScope)
if err != nil {
    // Handle error
}

// Initialize Search client
app.SearchClient = search.NewClient(accessToken)

// Perform search
searchResults, err := app.SearchClient.Search(ctx, searchQuery)
```

### User Interface

The example provides a simple web interface with several pages:

- Home page with authentication status
- User profile page
- Flows dashboard that shows a user's flows
- Search dashboard for searching Globus data

The interface uses vanilla JavaScript to create a responsive experience without dependencies.

## Running the Example

### Prerequisites

To run the example, you need:

1. Go 1.16 or later
2. A registered Globus app with:
   - Client ID and Client Secret
   - Redirect URL: `http://localhost:8080/callback`
   - Required scopes: openid, profile, email, flows, search

### Configuration

Set the following environment variables:

```bash
export GLOBUS_CLIENT_ID="your-client-id"
export GLOBUS_CLIENT_SECRET="your-client-secret"
export SESSION_SECRET="random-secret-for-sessions"
```

### Starting the Server

Navigate to the example directory and run:

```bash
cd examples/webapp
go run main.go
```

The application will start on port 8080. Open your browser to `http://localhost:8080` to use the application.

## Code Structure

The example follows a simple MVC pattern:

- **Model**: Session and token management
- **View**: HTML templates rendered server-side
- **Controller**: HTTP handlers for different routes

The main components:

1. **Main Function**: Sets up the application state and HTTP routes
2. **HTTP Handlers**: Functions that handle different routes
3. **API Handlers**: Functions that provide JSON APIs for the frontend
4. **Utility Functions**: Functions for session and token management

## Security Considerations

The example implements several security measures:

- CSRF protection using state parameters
- HttpOnly cookies
- Input validation
- Scope-specific token handling

However, for production use, you would want to add:

- HTTPS with proper certificates
- More robust session management
- Database storage for sessions and application state
- Rate limiting
- Additional input validation
- Proper error handling and logging

## Best Practices Demonstrated

The example demonstrates several best practices:

1. **Separation of Concerns**: Clear separation between authentication, session management, and service calls
2. **Token Handling**: Secure storage and refreshing of tokens
3. **Scope Management**: Using specific scopes for different services
4. **Error Handling**: Proper error handling for API calls
5. **Security Measures**: Implementing CSRF protection and other security measures

## Extending the Example

To extend the example:

1. **Add More Services**: Integrate with additional Globus services like Transfer
2. **Improve UI**: Add more interactive features and better styling
3. **Add Database**: Use a proper database for session storage
4. **Add User Management**: Implement user registration and management
5. **Add Logging**: Implement proper logging for monitoring and debugging
6. **Add Tests**: Add comprehensive tests for the application

## Common Issues

**Token Refresh Failures**:
- Ensure the refresh token is properly stored
- Verify the client ID and secret are correct
- Check the scopes requested during authentication

**Session Management Issues**:
- Clear cookies and session storage if sessions are not working
- Ensure cookies are properly set and read
- Check for session expiration

**Service Integration Issues**:
- Verify the access token has the correct scopes
- Check for API errors in the console
- Verify network connectivity to Globus services

## Conclusion

This example demonstrates how to use the Globus Go SDK in a web application, showcasing authentication, token management, and service integration. It provides a foundation for building more complex web applications that leverage Globus services.