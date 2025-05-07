// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: 2025 Scott Friedman and Project Contributors
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/scttfrdmn/globus-go-sdk/pkg/services/auth"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/flows"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/search"
	"github.com/scttfrdmn/globus-go-sdk/pkg/services/tokens"
)

// Configuration constants
const (
	serverPort        = "8080"
	sessionCookieName = "globus-session"
	tokensDir         = "./tokens"
)

// Application configuration
type Config struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

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
	Config        *Config
	Sessions      map[string]*Session
	TokenStorage  tokens.Storage
	TokenManager  *tokens.Manager
	AuthClient    *auth.Client
	FlowsClient   *flows.Client
	SearchClient  *search.Client
	SessionSecret []byte
}

func main() {
	// Load configuration from environment
	config := &Config{
		ClientID:     os.Getenv("GLOBUS_CLIENT_ID"),
		ClientSecret: os.Getenv("GLOBUS_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:" + serverPort + "/callback",
		Scopes: []string{
			auth.ScopeOpenID,
			auth.ScopeProfile,
			auth.ScopeEmail,
			flows.FlowsScope,
			search.SearchScope,
		},
	}

	// Validate configuration
	if config.ClientID == "" || config.ClientSecret == "" {
		log.Fatal("GLOBUS_CLIENT_ID and GLOBUS_CLIENT_SECRET environment variables are required")
	}

	// Create token storage directory if it doesn't exist
	if err := os.MkdirAll(tokensDir, 0700); err != nil {
		log.Fatalf("Failed to create token directory: %v", err)
	}

	// Initialize app state
	app := &App{
		Config:        config,
		Sessions:      make(map[string]*Session),
		SessionSecret: []byte(os.Getenv("SESSION_SECRET")), // In production, use a proper secret
	}

	// Initialize token storage
	var err error
	app.TokenStorage, err = tokens.NewFileStorage(tokensDir)
	if err != nil {
		log.Fatalf("Failed to initialize token storage: %v", err)
	}

	// Initialize auth client
	app.AuthClient = auth.NewClient(config.ClientID, config.ClientSecret)
	app.AuthClient.SetRedirectURL(config.RedirectURL)

	// Initialize token manager
	app.TokenManager = tokens.NewManager(app.TokenStorage, app.AuthClient)

	// Configure token refresh to happen when tokens are within 10 minutes of expiry
	app.TokenManager.SetRefreshThreshold(10 * time.Minute)

	// Start background token refresh (refresh every 15 minutes)
	stopRefresh := app.TokenManager.StartBackgroundRefresh(15 * time.Minute)
	defer stopRefresh() // Will be called when the app terminates

	// Set up HTTP routes
	http.HandleFunc("/", app.handleHome)
	http.HandleFunc("/login", app.handleLogin)
	http.HandleFunc("/callback", app.handleCallback)
	http.HandleFunc("/logout", app.handleLogout)
	http.HandleFunc("/profile", app.handleProfile)
	http.HandleFunc("/flows", app.handleFlows)
	http.HandleFunc("/search", app.handleSearch)
	http.HandleFunc("/api/flows", app.handleAPIFlows)
	http.HandleFunc("/api/search", app.handleAPISearch)

	// Start server
	fmt.Printf("Starting server on port %s...\n", serverPort)
	fmt.Printf("Open your browser to http://localhost:%s\n", serverPort)
	log.Fatal(http.ListenAndServe(":"+serverPort, nil))
}

// handleHome renders the home page
func (app *App) handleHome(w http.ResponseWriter, r *http.Request) {
	session := app.getSession(r)

	if session == nil {
		// Not logged in
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Globus SDK Web Example</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        h1 { color: #2971D6; }
        .btn { 
            display: inline-block; 
            background-color: #2971D6; 
            color: white; 
            padding: 10px 15px; 
            text-decoration: none; 
            border-radius: 4px; 
            margin: 10px 0;
        }
    </style>
</head>
<body>
    <h1>Globus SDK Web Example</h1>
    <p>This is a demonstration of using the Globus Go SDK in a web application.</p>
    <p>To get started, log in with your Globus account:</p>
    <a href="/login" class="btn">Log in with Globus</a>
</body>
</html>
		`)
		return
	}

	// User is logged in
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Globus SDK Web Example</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        h1, h2 { color: #2971D6; }
        .btn { 
            display: inline-block; 
            background-color: #2971D6; 
            color: white; 
            padding: 10px 15px; 
            text-decoration: none; 
            border-radius: 4px; 
            margin: 10px 5px 10px 0;
        }
        .btn-secondary {
            background-color: #666;
        }
        nav { margin: 20px 0; }
    </style>
</head>
<body>
    <h1>Globus SDK Web Example</h1>
    <p>Welcome, %s!</p>

    <nav>
        <a href="/profile" class="btn">View Profile</a>
        <a href="/flows" class="btn">Flows Dashboard</a>
        <a href="/search" class="btn">Search Dashboard</a>
        <a href="/logout" class="btn btn-secondary">Log Out</a>
    </nav>

    <h2>About This Example</h2>
    <p>This web application demonstrates several features of the Globus Go SDK:</p>
    <ul>
        <li>OAuth2 authentication flow with Globus</li>
        <li>Token storage and automatic refresh</li>
        <li>Flows service integration</li>
        <li>Search service integration</li>
    </ul>
</body>
</html>
	`, session.Username)
}

// handleLogin initiates the OAuth2 flow
func (app *App) handleLogin(w http.ResponseWriter, r *http.Request) {
	// Generate a state parameter to protect against CSRF
	state := fmt.Sprintf("%d", time.Now().UnixNano())

	// Create the authorization URL
	authURL := app.AuthClient.GetAuthorizationURL(
		state,
		app.Config.Scopes...,
	)

	// Store the state in a cookie for verification later
	http.SetCookie(w, &http.Cookie{
		Name:     "globus_auth_state",
		Value:    state,
		Path:     "/",
		Secure:   r.TLS != nil,
		HttpOnly: true,
		MaxAge:   300, // 5 minutes
	})

	// Redirect to Globus Auth
	http.Redirect(w, r, authURL, http.StatusFound)
}

// handleCallback processes the OAuth2 callback
func (app *App) handleCallback(w http.ResponseWriter, r *http.Request) {
	// Verify state parameter to prevent CSRF
	stateCookie, err := r.Cookie("globus_auth_state")
	if err != nil {
		http.Error(w, "Missing state cookie", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")
	if state != stateCookie.Value {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Clear the state cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "globus_auth_state",
		Value:    "",
		Path:     "/",
		Secure:   r.TLS != nil,
		HttpOnly: true,
		MaxAge:   -1,
	})

	// Get authorization code from query
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "Missing authorization code", http.StatusBadRequest)
		return
	}

	// Exchange code for tokens
	ctx := context.Background()
	tokenResponse, err := app.AuthClient.ExchangeAuthorizationCode(ctx, code)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to exchange code: %v", err), http.StatusInternalServerError)
		return
	}

	// Get user info
	userInfo, err := app.getUserInfo(ctx, tokenResponse.AccessToken)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get user info: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a new session
	sessionID := fmt.Sprintf("%d", time.Now().UnixNano())
	session := &Session{
		ID:           sessionID,
		UserID:       userInfo.Sub,
		Username:     userInfo.PreferredUsername,
		Email:        userInfo.Email,
		SessionStart: time.Now(),
		LastAccess:   time.Now(),
	}
	app.Sessions[sessionID] = session

	// Store tokens in token storage
	if err := app.storeTokens(session.UserID, tokenResponse); err != nil {
		http.Error(w, fmt.Sprintf("Failed to store tokens: %v", err), http.StatusInternalServerError)
		return
	}

	// Set session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		Secure:   r.TLS != nil,
		HttpOnly: true,
		MaxAge:   3600 * 24, // 24 hours
	})

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusFound)
}

// handleLogout handles user logout
func (app *App) handleLogout(w http.ResponseWriter, r *http.Request) {
	// Get session
	session := app.getSession(r)
	if session == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Remove session
	delete(app.Sessions, session.ID)

	// Clear session cookie
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		Secure:   r.TLS != nil,
		HttpOnly: true,
		MaxAge:   -1,
	})

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusFound)
}

// handleProfile displays the user profile
func (app *App) handleProfile(w http.ResponseWriter, r *http.Request) {
	session := app.getSession(r)
	if session == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Profile - Globus SDK Web Example</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        h1, h2 { color: #2971D6; }
        .btn { 
            display: inline-block; 
            background-color: #2971D6; 
            color: white; 
            padding: 10px 15px; 
            text-decoration: none; 
            border-radius: 4px; 
            margin: 10px 0;
        }
        nav { margin: 20px 0; }
        table { width: 100%%; border-collapse: collapse; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <h1>User Profile</h1>
    <nav>
        <a href="/" class="btn">Home</a>
    </nav>

    <h2>Account Information</h2>
    <table>
        <tr>
            <th>Username</th>
            <td>%s</td>
        </tr>
        <tr>
            <th>Email</th>
            <td>%s</td>
        </tr>
        <tr>
            <th>User ID</th>
            <td>%s</td>
        </tr>
        <tr>
            <th>Session Start</th>
            <td>%s</td>
        </tr>
        <tr>
            <th>Last Access</th>
            <td>%s</td>
        </tr>
    </table>
</body>
</html>
	`, session.Username, session.Email, session.UserID,
		session.SessionStart.Format(time.RFC1123),
		session.LastAccess.Format(time.RFC1123))
}

// handleFlows displays the flows dashboard
func (app *App) handleFlows(w http.ResponseWriter, r *http.Request) {
	session := app.getSession(r)
	if session == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Flows Dashboard - Globus SDK Web Example</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        h1, h2 { color: #2971D6; }
        .btn { 
            display: inline-block; 
            background-color: #2971D6; 
            color: white; 
            padding: 10px 15px; 
            text-decoration: none; 
            border-radius: 4px; 
            margin: 10px 5px 10px 0;
        }
        nav { margin: 20px 0; }
        #flowsList { margin-top: 20px; }
        .loading { color: #666; font-style: italic; }
        .error { color: red; }
        table { width: 100%%; border-collapse: collapse; margin-top: 20px; }
        th, td { padding: 10px; text-align: left; border-bottom: 1px solid #ddd; }
        th { background-color: #f2f2f2; }
        .flow-item { margin-bottom: 10px; padding: 10px; border: 1px solid #ddd; border-radius: 4px; }
        .flow-item h3 { margin-top: 0; }
    </style>
    <script>
        // Fetch flows when the page loads
        document.addEventListener('DOMContentLoaded', function() {
            fetchFlows();
        });

        // Fetch flows from the API
        function fetchFlows() {
            const flowsList = document.getElementById('flowsList');
            flowsList.innerHTML = '<p class="loading">Loading flows...</p>';
            
            fetch('/api/flows')
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Network response was not ok');
                    }
                    return response.json();
                })
                .then(data => {
                    if (data.flows && data.flows.length > 0) {
                        let html = '<table>';
                        html += '<tr><th>Title</th><th>Owner</th><th>Created</th><th>Public</th></tr>';
                        
                        data.flows.forEach(flow => {
                            html += '<tr>';
                            html += '<td>' + escapeHtml(flow.title) + '</td>';
                            html += '<td>' + escapeHtml(flow.flow_owner || 'N/A') + '</td>';
                            html += '<td>' + formatDate(flow.created_at) + '</td>';
                            html += '<td>' + (flow.public ? 'Yes' : 'No') + '</td>';
                            html += '</tr>';
                        });
                        
                        html += '</table>';
                        flowsList.innerHTML = html;
                    } else {
                        flowsList.innerHTML = '<p>No flows found. You may need to create flows in the Globus web interface.</p>';
                    }
                })
                .catch(error => {
                    console.error('Error fetching flows:', error);
                    flowsList.innerHTML = '<p class="error">Error loading flows: ' + error.message + '</p>';
                });
        }
        
        // Helper function to escape HTML
        function escapeHtml(str) {
            if (!str) return '';
            return str
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;')
                .replace(/"/g, '&quot;')
                .replace(/'/g, '&#039;');
        }
        
        // Format date
        function formatDate(dateStr) {
            if (!dateStr) return 'N/A';
            const date = new Date(dateStr);
            return date.toLocaleString();
        }
    </script>
</head>
<body>
    <h1>Flows Dashboard</h1>
    <nav>
        <a href="/" class="btn">Home</a>
    </nav>

    <h2>Your Flows</h2>
    <p>This dashboard displays your Globus Flows:</p>
    
    <div id="flowsList">
        <p class="loading">Loading flows...</p>
    </div>
</body>
</html>
	`)
}

// handleSearch displays the search dashboard
func (app *App) handleSearch(w http.ResponseWriter, r *http.Request) {
	session := app.getSession(r)
	if session == nil {
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, `
<!DOCTYPE html>
<html>
<head>
    <title>Search Dashboard - Globus SDK Web Example</title>
    <style>
        body { font-family: Arial, sans-serif; max-width: 800px; margin: 0 auto; padding: 20px; }
        h1, h2 { color: #2971D6; }
        .btn { 
            display: inline-block; 
            background-color: #2971D6; 
            color: white; 
            padding: 10px 15px; 
            text-decoration: none; 
            border-radius: 4px; 
            margin: 10px 5px 10px 0;
        }
        nav { margin: 20px 0; }
        form { margin: 20px 0; }
        input[type="text"] { 
            padding: 8px; 
            width: 70%%; 
            font-size: 16px; 
        }
        button { 
            padding: 8px 15px; 
            background-color: #2971D6; 
            color: white; 
            border: none; 
            font-size: 16px; 
            cursor: pointer; 
        }
        #searchResults { margin-top: 20px; }
        .loading { color: #666; font-style: italic; }
        .error { color: red; }
        .result-item { 
            margin-bottom: 15px; 
            padding: 10px; 
            border: 1px solid #ddd; 
            border-radius: 4px; 
        }
        .result-item h3 { margin-top: 0; }
    </style>
    <script>
        // Handle form submission
        document.addEventListener('DOMContentLoaded', function() {
            const searchForm = document.getElementById('searchForm');
            searchForm.addEventListener('submit', function(event) {
                event.preventDefault();
                const query = document.getElementById('searchQuery').value;
                if (query.trim()) {
                    performSearch(query);
                }
            });
        });

        // Perform search
        function performSearch(query) {
            const searchResults = document.getElementById('searchResults');
            searchResults.innerHTML = '<p class="loading">Searching...</p>';
            
            fetch('/api/search?q=' + encodeURIComponent(query))
                .then(response => {
                    if (!response.ok) {
                        throw new Error('Network response was not ok');
                    }
                    return response.json();
                })
                .then(data => {
                    if (data.results && data.results.gmeta) {
                        const items = data.results.gmeta;
                        if (items.length > 0) {
                            let html = '<h3>Found ' + items.length + ' results:</h3>';
                            
                            items.forEach(item => {
                                html += '<div class="result-item">';
                                html += '<h3>' + escapeHtml(item.subject || 'Untitled') + '</h3>';
                                
                                // Display visible metadata
                                if (item.visible_to) {
                                    html += '<p><strong>Visible to:</strong> ' + escapeHtml(item.visible_to.join(', ')) + '</p>';
                                }
                                
                                // Display content if available
                                if (item.content) {
                                    for (const key in item.content) {
                                        if (typeof item.content[key] === 'object') {
                                            html += '<p><strong>' + escapeHtml(key) + ':</strong> ' + 
                                                    JSON.stringify(item.content[key]) + '</p>';
                                        } else {
                                            html += '<p><strong>' + escapeHtml(key) + ':</strong> ' + 
                                                    escapeHtml(String(item.content[key])) + '</p>';
                                        }
                                    }
                                }
                                
                                html += '</div>';
                            });
                            
                            searchResults.innerHTML = html;
                        } else {
                            searchResults.innerHTML = '<p>No results found for your search.</p>';
                        }
                    } else {
                        searchResults.innerHTML = '<p>No results found for your search.</p>';
                    }
                })
                .catch(error => {
                    console.error('Error performing search:', error);
                    searchResults.innerHTML = '<p class="error">Error searching: ' + error.message + '</p>';
                });
        }
        
        // Helper function to escape HTML
        function escapeHtml(str) {
            if (!str) return '';
            return String(str)
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;')
                .replace(/"/g, '&quot;')
                .replace(/'/g, '&#039;');
        }
    </script>
</head>
<body>
    <h1>Search Dashboard</h1>
    <nav>
        <a href="/" class="btn">Home</a>
    </nav>

    <h2>Globus Search</h2>
    <p>Search for data in Globus Search:</p>
    
    <form id="searchForm">
        <input type="text" id="searchQuery" placeholder="Enter search query" required>
        <button type="submit">Search</button>
    </form>
    
    <div id="searchResults">
        <!-- Search results will appear here -->
    </div>
</body>
</html>
	`)
}

// handleAPIFlows returns flows data as JSON
func (app *App) handleAPIFlows(w http.ResponseWriter, r *http.Request) {
	session := app.getSession(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get tokens for the user
	ctx := context.Background()
	accessToken, err := app.getAccessToken(ctx, session.UserID, flows.FlowsScope)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get access token: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a new flows client with the fresh token for this request
	app.FlowsClient = flows.NewClient(accessToken)

	// Note: In a production implementation, we would use the flows client's method directly
	// but for this example we're using our own implementation

	// For this example, we'll use our own simple flows implementation
	flowList, err := app.performSimpleFlowsList(ctx, 10)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list flows: %v", err), http.StatusInternalServerError)
		return
	}

	// Return flows as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(flowList)
}

// handleAPISearch handles search API requests
func (app *App) handleAPISearch(w http.ResponseWriter, r *http.Request) {
	session := app.getSession(r)
	if session == nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get query parameter
	query := r.URL.Query().Get("q")
	if query == "" {
		http.Error(w, "Missing query parameter", http.StatusBadRequest)
		return
	}

	// Get tokens for the user
	ctx := context.Background()
	accessToken, err := app.getAccessToken(ctx, session.UserID, search.SearchScope)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get access token: %v", err), http.StatusInternalServerError)
		return
	}

	// Create a new search client with the fresh token for this request
	app.SearchClient = search.NewClient(accessToken)

	// Note: In a production implementation, we would use the search client's method directly
	// but for this example we're using our own implementation

	// For this example, we'll use our own simple search implementation
	searchResults, err := app.performSimpleSearch(ctx, query, 10)
	if err != nil {
		http.Error(w, fmt.Sprintf("Search failed: %v", err), http.StatusInternalServerError)
		return
	}

	// Return results as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"query":   query,
		"results": searchResults,
	})
}

// getSession retrieves the current session from the request
func (app *App) getSession(r *http.Request) *Session {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return nil
	}

	session, ok := app.Sessions[cookie.Value]
	if !ok {
		return nil
	}

	// Update last access time
	session.LastAccess = time.Now()
	return session
}

// storeTokens stores tokens for a user
func (app *App) storeTokens(userID string, tokenResponse *auth.TokenResponse) error {
	// Create a token entry with TokenSet
	entry := &tokens.Entry{
		Resource: userID,
		TokenSet: &tokens.TokenSet{
			AccessToken:  tokenResponse.AccessToken,
			RefreshToken: tokenResponse.RefreshToken,
			ExpiresAt:    time.Now().Add(time.Duration(tokenResponse.ExpiresIn) * time.Second),
			Scope:        tokenResponse.Scope,
		},
	}

	// Store the tokens
	return app.TokenStorage.Store(entry)
}

// getAccessToken gets an access token for the specified user and scope
func (app *App) getAccessToken(ctx context.Context, userID, scope string) (string, error) {
	// Use the TokenManager to get (and potentially refresh) the token
	entry, err := app.TokenManager.GetToken(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w", err)
	}

	// Check if token has the required scope
	if entry.TokenSet.Scope == "" || containsScope(entry.TokenSet.Scope, scope) {
		return entry.TokenSet.AccessToken, nil
	}

	return "", fmt.Errorf("token does not have required scope %s", scope)
}

// getUserInfo retrieves user information from Globus Auth
func (app *App) getUserInfo(ctx context.Context, accessToken string) (*auth.UserInfo, error) {
	return app.AuthClient.GetUserInfo(ctx, accessToken)
}

// containsScope checks if a scope string contains a specific scope
func containsScope(scopeStr, scope string) bool {
	// Simple check - should be improved with proper scope parsing
	return scopeStr == scope || scopeStr == "all" || scopeStr == "*"
}
