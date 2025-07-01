package console

// SystemInfo represents a system in the catalog
type SystemInfo struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Category    string   `json:"category"`
	Difficulty  string   `json:"difficulty"`
	Tags        []string `json:"tags"`
	Thumbnail   string   `json:"thumbnail,omitempty"`
	LastUpdated string   `json:"lastUpdated"`
}

// SystemProject represents a full system project
type SystemProject struct {
	ID             string                   `json:"id"`
	Name           string                   `json:"name"`
	Description    string                   `json:"description"`
	Category       string                   `json:"category"`
	Difficulty     string                   `json:"difficulty"`
	Tags           []string                 `json:"tags"`
	Versions       map[string]SystemVersion `json:"versions"`
	DefaultVersion string                   `json:"defaultVersion"`
	LastUpdated    string                   `json:"lastUpdated"`
}

// SystemVersion represents a version of a system
type SystemVersion struct {
	SDL    string `json:"sdl"`
	Recipe string `json:"recipe"`
	Readme string `json:"readme,omitempty"`
}

// SystemCatalogService manages the system examples catalog
type SystemCatalogService struct {
	systems map[string]*SystemProject
}

// NewSystemCatalogService creates a new system catalog service
func NewSystemCatalogService() *SystemCatalogService {
	service := &SystemCatalogService{
		systems: make(map[string]*SystemProject),
	}
	service.initializeCatalog()
	return service
}

// initializeCatalog loads the example systems
func (s *SystemCatalogService) initializeCatalog() {
	// Bitly URL Shortener
	s.systems["bitly"] = &SystemProject{
		ID:             "bitly",
		Name:           "Bitly URL Shortener",
		Description:    "A scalable URL shortening service with analytics and caching",
		Category:       "Web Services",
		Difficulty:     "beginner",
		Tags:           []string{"web", "database", "caching", "rest-api"},
		DefaultVersion: "v1",
		LastUpdated:    "2024-01-15T10:00:00Z",
		Versions: map[string]SystemVersion{
			"v1": {
				SDL: `// Bitly URL Shortener System
// A simple URL shortening service with caching

import stdlib.storage.PostgreSQL
import stdlib.cache.Redis
import stdlib.compute.LoadBalancer

system Bitly {
  // API Gateway handles incoming requests
  gateway := LoadBalancer {
    port: 443
    protocol: "HTTPS"
    healthCheck: "/health"
  }
  
  // URL shortening service
  service URLShortener {
    port: 8080
    replicas: 3
    
    endpoints: {
      POST "/shorten": "Create short URL"
      GET "/:short": "Redirect to long URL"
      GET "/stats/:short": "Get URL statistics"
    }
  }
  
  // PostgreSQL for persistent storage
  database := PostgreSQL {
    version: "14"
    storage: "100GB"
    replicas: {
      primary: 1
      read: 2
    }
  }
  
  // Redis for caching popular URLs
  cache := Redis {
    memory: "16GB"
    eviction: "LRU"
    persistence: true
  }
  
  // Connect components
  gateway -> URLShortener
  URLShortener -> database
  URLShortener -> cache
}`,
				Recipe: `# Bitly System Demo Recipe

echo "Starting Bitly URL Shortener demo..."
echo "This demo will show URL shortening and redirection"

# Start the system
start system
pause

# Generate some test traffic
echo "Creating short URLs..."
sdl traffic http POST /shorten '{"url": "https://example.com/very/long/url/path"}' --rate 5/s --duration 20s

pause

echo "Testing URL redirects..."
sdl traffic http GET /abc123 --rate 20/s --duration 20s

# Show metrics
echo "Starting metrics collection..."
start metrics

pause

echo "Demo complete! The system handled URL shortening and redirections."
stop`,
			},
		},
	}

	// Uber Ride Sharing (Basic)
	s.systems["uber-basic"] = &SystemProject{
		ID:             "uber-basic",
		Name:           "Uber Ride Sharing (Basic)",
		Description:    "Simplified ride-sharing platform with driver matching",
		Category:       "Transportation",
		Difficulty:     "intermediate",
		Tags:           []string{"microservices", "real-time", "geo-spatial", "matching"},
		DefaultVersion: "v1",
		LastUpdated:    "2024-02-10T14:30:00Z",
		Versions: map[string]SystemVersion{
			"v1": {
				SDL: `// Uber Ride Sharing System (Simplified)
// Basic ride matching and tracking

import stdlib.storage.PostgreSQL
import stdlib.cache.Redis
import stdlib.compute.LoadBalancer
import stdlib.messaging.Kafka

system UberBasic {
  // API Gateway
  gateway := LoadBalancer {
    port: 443
    protocol: "HTTPS"
  }
  
  // User Service
  service UserService {
    port: 8080
    replicas: 2
    
    endpoints: {
      POST "/users": "Register user"
      GET "/users/:id": "Get user info"
      POST "/rides/request": "Request a ride"
    }
  }
  
  // Driver Service
  service DriverService {
    port: 8081
    replicas: 2
    
    endpoints: {
      POST "/drivers": "Register driver"
      POST "/drivers/:id/location": "Update location"
      POST "/drivers/:id/status": "Update availability"
    }
  }
  
  // Matching Service
  service MatchingService {
    port: 8082
    replicas: 3
    
    endpoints: {
      POST "/match": "Match rider with driver"
      GET "/matches/:id": "Get match details"
    }
  }
  
  // Location tracking cache
  locationCache := Redis {
    memory: "32GB"
    eviction: "LRU"
  }
  
  // Main database
  database := PostgreSQL {
    version: "14"
    storage: "500GB"
  }
  
  // Event streaming
  eventBus := Kafka {
    partitions: 10
    replication: 3
  }
  
  // Connections
  gateway -> UserService
  gateway -> DriverService
  gateway -> MatchingService
  
  UserService -> database
  UserService -> eventBus
  
  DriverService -> database
  DriverService -> locationCache
  DriverService -> eventBus
  
  MatchingService -> locationCache
  MatchingService -> eventBus
}`,
				Recipe: `# Uber Ride Sharing Demo

echo "Starting Uber ride sharing simulation..."

# Start the system
start system
pause

echo "Registering drivers..."
sdl traffic http POST /drivers '{"name": "Driver {{seq}}", "vehicle": "Toyota Camry"}' --rate 2/s --duration 10s

pause

echo "Updating driver locations..."
sdl traffic http POST /drivers/{{random:1-20}}/location '{"lat": {{random:37.7-37.8}}, "lng": {{random:-122.5--122.4}}}' --rate 10/s --duration 30s

pause

echo "Simulating ride requests..."
sdl traffic http POST /rides/request '{"pickup": {"lat": {{random:37.7-37.8}}, "lng": {{random:-122.5--122.4}}}, "destination": {"lat": {{random:37.7-37.8}}, "lng": {{random:-122.5--122.4}}}}' --rate 5/s --duration 20s

# Show metrics
start metrics

pause

echo "Demo complete! The system processed ride requests and driver updates."
stop`,
			},
		},
	}

	// Chat Application
	s.systems["chat-app"] = &SystemProject{
		ID:             "chat-app",
		Name:           "Real-time Chat Application",
		Description:    "WebSocket-based chat with message history",
		Category:       "Communication",
		Difficulty:     "beginner",
		Tags:           []string{"websocket", "real-time", "messaging"},
		DefaultVersion: "v1",
		LastUpdated:    "2024-03-05T09:15:00Z",
		Versions: map[string]SystemVersion{
			"v1": {
				SDL: `// Real-time Chat Application
// Simple chat system with WebSocket support

import stdlib.storage.PostgreSQL
import stdlib.cache.Redis
import stdlib.compute.LoadBalancer

system ChatApp {
  // WebSocket Gateway
  gateway := LoadBalancer {
    port: 443
    protocol: "WSS"
    sticky: true  // Sticky sessions for WebSocket
  }
  
  // Chat Service
  service ChatService {
    port: 8080
    replicas: 4
    
    endpoints: {
      WS "/chat": "WebSocket chat connection"
      POST "/rooms": "Create chat room"
      GET "/rooms/:id/history": "Get message history"
    }
  }
  
  // Message store
  database := PostgreSQL {
    version: "14"
    storage: "200GB"
  }
  
  // Active connections cache
  cache := Redis {
    memory: "8GB"
    pubsub: true  // For real-time message broadcasting
  }
  
  // Connections
  gateway -> ChatService
  ChatService -> database
  ChatService -> cache
}`,
				Recipe: `# Chat Application Demo

echo "Starting real-time chat demo..."

# Start the system
start system
pause

echo "Creating chat rooms..."
sdl traffic http POST /rooms '{"name": "Room {{seq}}", "topic": "General Chat"}' --rate 1/s --duration 5s

pause

echo "Simulating chat messages..."
sdl traffic ws /chat '{"room": "room-1", "message": "Hello from user {{seq}}!", "user": "user-{{seq}}"}' --connections 50 --rate 2/s --duration 30s

# Show metrics
start metrics

pause

echo "Demo complete! The chat system handled real-time messaging."
stop`,
			},
		},
	}
}

// ListSystems returns all systems as SystemInfo
func (s *SystemCatalogService) ListSystems() []SystemInfo {
	var systems []SystemInfo
	for _, project := range s.systems {
		systems = append(systems, SystemInfo{
			ID:          project.ID,
			Name:        project.Name,
			Description: project.Description,
			Category:    project.Category,
			Difficulty:  project.Difficulty,
			Tags:        project.Tags,
			LastUpdated: project.LastUpdated,
		})
	}
	return systems
}

// GetSystem returns a specific system project
func (s *SystemCatalogService) GetSystem(id string) *SystemProject {
	return s.systems[id]
}