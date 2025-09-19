package user

// Using a simple struct for domain entity
type User struct {
	Username string
	Password string // In a real app, this would be a hash
}

// UserService provides user-related operations.
type UserService struct {
	// Using a simple in-memory map for the MVP.
	// In a real application, this would be a UserRepository interface
	// with a concrete implementation (e.g., PostgreSQL).
	users map[string]string
}

// NewUserService creates a new UserService.
func NewUserService(users map[string]string) *UserService {
	return &UserService{users: users}
}

// Authenticate checks if a username and password are valid.
func (s *UserService) Authenticate(username, password string) bool {
	p, ok := s.users[username]
	if !ok {
		return false
	}
	return p == password
}
