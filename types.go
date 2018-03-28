package syshealth

// Data stores metrics identified by key
type Data map[string]interface{}

// MetricBag is a container used to transport metrics and eventually some metadata
type MetricBag struct {
	Metrics Data `json:"metrics"`
}

// Server represents server data
type Server struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	IP   string `json:"ip"`
}

// ServerRepository defines the behaviour of the server repository
type ServerRepository interface {
	// GetServers returns the complete list of registered servers
	GetServers() ([]Server, error)
	// RegisterServer returns a JWT or an error
	RegisterServer(server Server, jwtSecret string) (string, error)
	// RevokeServer revokes a server token
	RevokeServer(id string) error
	// CheckServerIsRegistered checks if a server is registered
	CheckServerIsRegistered(id string) (bool, error)
}

// MetricRepository defines the behaviour of the metric repository
type MetricRepository interface {
	Get(serverID string) (*Data, error)
	Store(serverID string, data Data) error
}

// AdminUserRepository defines the behaviour of the admin user repository
type AdminUserRepository interface {
	IsSetup() (bool, error)
	Login(username string, password string) (bool, error)
	GetUsers() ([]string, error)
	Create(username string, password string) error
	Delete(username string) error
}
