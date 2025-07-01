/**
 * Service for managing system catalog and examples
 * This will eventually be replaced with server API calls
 */

export interface SystemProject {
  id: string;
  name: string;
  description: string;
  category: string;
  difficulty: 'beginner' | 'intermediate' | 'advanced';
  tags: string[];
  thumbnail?: string;
  versions: {
    [key: string]: {
      sdl: string;
      recipe: string;
      readme?: string;
    }
  };
  defaultVersion: string;
}

export class SystemCatalogService {
  private systems: Map<string, SystemProject> = new Map();

  constructor() {
    this.initializeCatalog();
  }

  private initializeCatalog(): void {
    // Initialize with example systems
    this.addSystem({
      id: 'bitly',
      name: 'Bitly URL Shortener',
      description: 'A scalable URL shortening service with analytics and caching',
      category: 'Web Services',
      difficulty: 'beginner',
      tags: ['web', 'database', 'caching', 'rest-api'],
      defaultVersion: 'v1',
      versions: {
        v1: {
          sdl: `// Bitly URL Shortener System
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
          recipe: `# Bitly System Demo Recipe

echo "🔗 Starting Bitly URL Shortener demo..."
echo "This demo will show URL shortening and redirection"

# Start the system
start system
pause

# Generate some test traffic
echo "📝 Creating short URLs..."
sdl traffic http POST /shorten '{"url": "https://example.com/very/long/url/path"}' --rate 5/s --duration 20s

pause

echo "🔄 Testing URL redirects..."
sdl traffic http GET /abc123 --rate 20/s --duration 20s

# Show metrics
echo "📊 Starting metrics collection..."
start metrics

pause

echo "✅ Demo complete! The system handled URL shortening and redirections."
stop`
        }
      }
    });

    this.addSystem({
      id: 'uber-basic',
      name: 'Uber Ride Sharing (Basic)',
      description: 'Simplified ride-sharing platform with driver matching',
      category: 'Transportation',
      difficulty: 'intermediate',
      tags: ['microservices', 'real-time', 'geo-spatial', 'matching'],
      defaultVersion: 'v1',
      versions: {
        v1: {
          sdl: `// Uber Ride Sharing System (Simplified)
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
          recipe: `# Uber Ride Sharing Demo

echo "🚗 Starting Uber ride sharing simulation..."

# Start the system
start system
pause

echo "👤 Registering drivers..."
sdl traffic http POST /drivers '{"name": "Driver {{seq}}", "vehicle": "Toyota Camry"}' --rate 2/s --duration 10s

pause

echo "📍 Updating driver locations..."
sdl traffic http POST /drivers/{{random:1-20}}/location '{"lat": {{random:37.7-37.8}}, "lng": {{random:-122.5--122.4}}}' --rate 10/s --duration 30s

pause

echo "🙋 Simulating ride requests..."
sdl traffic http POST /rides/request '{"pickup": {"lat": {{random:37.7-37.8}}, "lng": {{random:-122.5--122.4}}}, "destination": {"lat": {{random:37.7-37.8}}, "lng": {{random:-122.5--122.4}}}}' --rate 5/s --duration 20s

# Show metrics
start metrics

pause

echo "✅ Demo complete! The system processed ride requests and driver updates."
stop`
        }
      }
    });

    this.addSystem({
      id: 'chat-app',
      name: 'Real-time Chat Application',
      description: 'WebSocket-based chat with message history',
      category: 'Communication',
      difficulty: 'beginner',
      tags: ['websocket', 'real-time', 'messaging'],
      defaultVersion: 'v1',
      versions: {
        v1: {
          sdl: `// Real-time Chat Application
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
          recipe: `# Chat Application Demo

echo "💬 Starting real-time chat demo..."

# Start the system
start system
pause

echo "🏠 Creating chat rooms..."
sdl traffic http POST /rooms '{"name": "Room {{seq}}", "topic": "General Chat"}' --rate 1/s --duration 5s

pause

echo "💭 Simulating chat messages..."
sdl traffic ws /chat '{"room": "room-1", "message": "Hello from user {{seq}}!", "user": "user-{{seq}}"}' --connections 50 --rate 2/s --duration 30s

# Show metrics
start metrics

pause

echo "✅ Demo complete! The chat system handled real-time messaging."
stop`
        }
      }
    });

    this.addSystem({
      id: 'ecommerce',
      name: 'E-commerce Platform',
      description: 'Online shopping platform with inventory and orders',
      category: 'E-commerce',
      difficulty: 'intermediate',
      tags: ['web', 'database', 'inventory', 'payments'],
      defaultVersion: 'v1',
      versions: {
        v1: {
          sdl: `// E-commerce Platform
// Online shopping with inventory management

import stdlib.storage.PostgreSQL
import stdlib.cache.Redis
import stdlib.compute.LoadBalancer
import stdlib.messaging.RabbitMQ

system Ecommerce {
  // API Gateway
  gateway := LoadBalancer {
    port: 443
    protocol: "HTTPS"
    rateLimit: 1000  // requests per second
  }
  
  // Product Catalog Service
  service CatalogService {
    port: 8080
    replicas: 3
    
    endpoints: {
      GET "/products": "List products"
      GET "/products/:id": "Get product details"
      GET "/search": "Search products"
    }
  }
  
  // Shopping Cart Service
  service CartService {
    port: 8081
    replicas: 2
    
    endpoints: {
      POST "/cart/add": "Add to cart"
      DELETE "/cart/remove": "Remove from cart"
      GET "/cart": "View cart"
    }
  }
  
  // Order Service
  service OrderService {
    port: 8082
    replicas: 2
    
    endpoints: {
      POST "/orders": "Create order"
      GET "/orders/:id": "Get order status"
      POST "/orders/:id/cancel": "Cancel order"
    }
  }
  
  // Inventory Service
  service InventoryService {
    port: 8083
    replicas: 2
    
    endpoints: {
      POST "/inventory/reserve": "Reserve items"
      POST "/inventory/release": "Release reservation"
      GET "/inventory/:productId": "Check availability"
    }
  }
  
  // Main database
  database := PostgreSQL {
    version: "14"
    storage: "1TB"
    replicas: {
      primary: 1
      read: 2
    }
  }
  
  // Session cache
  sessionCache := Redis {
    memory: "16GB"
    ttl: 3600  // 1 hour session timeout
  }
  
  // Product cache
  productCache := Redis {
    memory: "32GB"
    eviction: "LRU"
  }
  
  // Order processing queue
  orderQueue := RabbitMQ {
    queues: ["orders", "inventory", "notifications"]
  }
  
  // Connections
  gateway -> CatalogService
  gateway -> CartService
  gateway -> OrderService
  
  CatalogService -> database
  CatalogService -> productCache
  
  CartService -> sessionCache
  CartService -> InventoryService
  
  OrderService -> database
  OrderService -> orderQueue
  OrderService -> InventoryService
  
  InventoryService -> database
  InventoryService -> orderQueue
}`,
          recipe: `# E-commerce Platform Demo

echo "🛍️ Starting e-commerce platform demo..."

# Start the system
start system
pause

echo "📦 Browsing product catalog..."
sdl traffic http GET /products?page={{random:1-10}} --rate 20/s --duration 15s

pause

echo "🔍 Searching for products..."
sdl traffic http GET /search?q={{random:laptop,phone,tablet,camera}} --rate 10/s --duration 15s

pause

echo "🛒 Adding items to cart..."
sdl traffic http POST /cart/add '{"productId": "prod-{{random:1-100}}", "quantity": {{random:1-3}}}' --rate 5/s --duration 20s

pause

echo "💳 Creating orders..."
sdl traffic http POST /orders '{"items": [{"productId": "prod-{{random:1-100}}", "quantity": {{random:1-2}}}], "payment": {"method": "credit_card"}}' --rate 2/s --duration 15s

# Show metrics
start metrics

pause

echo "✅ Demo complete! The e-commerce platform processed browsing, cart, and order operations."
stop`
        }
      }
    });
  }

  private addSystem(system: SystemProject): void {
    this.systems.set(system.id, system);
  }

  async listSystems(): Promise<SystemProject[]> {
    // In real implementation, this would be an API call
    return Array.from(this.systems.values());
  }

  async getSystem(id: string): Promise<SystemProject | null> {
    // In real implementation, this would be an API call
    return this.systems.get(id) || null;
  }

  async getSystemVersion(id: string, version: string): Promise<{ sdl: string; recipe: string } | null> {
    const system = this.systems.get(id);
    if (!system) return null;
    
    const versionData = system.versions[version];
    if (!versionData) {
      // Try default version
      return system.versions[system.defaultVersion] || null;
    }
    
    return versionData;
  }

  async searchSystems(query: string): Promise<SystemProject[]> {
    const lowerQuery = query.toLowerCase();
    return Array.from(this.systems.values()).filter(system => 
      system.name.toLowerCase().includes(lowerQuery) ||
      system.description.toLowerCase().includes(lowerQuery) ||
      system.tags.some(tag => tag.toLowerCase().includes(lowerQuery))
    );
  }

  async getSystemsByDifficulty(difficulty: string): Promise<SystemProject[]> {
    if (difficulty === 'all') {
      return this.listSystems();
    }
    
    return Array.from(this.systems.values()).filter(system => 
      system.difficulty === difficulty
    );
  }
}

// Singleton instance
export const systemCatalog = new SystemCatalogService();