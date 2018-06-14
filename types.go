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
	// GetServer returns the server associated to the given id, if it is registered
	GetServer(id string) (*Server, error)
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

// WatcherData represents data needed by watchers
type WatcherData struct {
	Server  Server
	Metrics Data
}

// Alert represents data for sending alert
type Alert struct {
	IssueTitle string
	Server     Server
	Level      ThresholdLevel
}

// ThresholdLevel represents a level of threshold
type ThresholdLevel int

const (
	// None represents that everything is OK
	None ThresholdLevel = 0
	// Warning represents the warning threshold
	Warning ThresholdLevel = 1
	// Critical represents the critical threshold
	Critical ThresholdLevel = 2
)

// Label returns a label representing the level
func (l ThresholdLevel) Label() string {
	switch l {
	case Critical:
		return "Critical"
	case Warning:
		return "Warning"
	default:
		return ""
	}
}

// WatcherKey represents a key to identify a watcher
type WatcherKey string

// Watcher defines the behaviour for watching metrics data
type Watcher interface {
	GetKey() WatcherKey
	Watch(data WatcherData)
}
