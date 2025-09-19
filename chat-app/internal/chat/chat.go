package chat

import (
	"log"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients. Mapped by username for easy lookup.
	Clients map[string]*Client

	// Inbound messages from the clients.
	Broadcast chan *Message

	// Register requests from the clients.
	Register chan *Client

	// Unregister requests from clients.
	Unregister chan *Client

	// Group management.
	Groups map[string]map[*Client]bool
}

func NewHub() *Hub {
	return &Hub{
		Broadcast:  make(chan *Message),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		Clients:    make(map[string]*Client),
		Groups:     make(map[string]map[*Client]bool),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.registerClient(client)
		case client := <-h.Unregister:
			h.unregisterClient(client)
		case message := <-h.Broadcast:
			h.handleMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	log.Printf("Registering client: %s", client.Username)
	h.Clients[client.Username] = client
	h.broadcastUserList()
}

func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.Clients[client.Username]; ok {
		log.Printf("Unregistering client: %s", client.Username)
		delete(h.Clients, client.Username)
		// Remove from all groups
		for groupName := range h.Groups {
			delete(h.Groups[groupName], client)
		}
		close(client.Send)
		h.broadcastUserList()
	}
}

func (h *Hub) handleMessage(message *Message) {
	log.Printf("Handling message: Type=%s, To=%s, From=%s", message.Type, message.Recipient, message.Sender)
	switch message.Type {
	case "private_message":
		if recipient, ok := h.Clients[message.Recipient]; ok {
			select {
			case recipient.Send <- message:
			default:
				close(recipient.Send)
				delete(h.Clients, recipient.Username)
			}
		}
	case "join_group":
		content, _ := message.Content.(string)
		h.joinGroup(message.Sender, content)
	case "leave_group":
		content, _ := message.Content.(string)
		h.leaveGroup(message.Sender, content)
	case "group_message":
		if members, ok := h.Groups[message.Recipient]; ok {
			senderClient := h.Clients[message.Sender]
			for client := range members {
				if client != senderClient { // Don't send to self
					select {
					case client.Send <- message:
					default:
						close(client.Send)
						delete(h.Clients, client.Username)
						delete(members, client)
					}
				}
			}
		}
	}
}

func (h *Hub) joinGroup(username string, groupName string) {
	client, ok := h.Clients[username]
	if !ok {
		return
	}
	if _, ok := h.Groups[groupName]; !ok {
		h.Groups[groupName] = make(map[*Client]bool)
	}
	h.Groups[groupName][client] = true
	log.Printf("User %s joined group %s", username, groupName)
}

func (h *Hub) leaveGroup(username string, groupName string) {
	client, ok := h.Clients[username]
	if !ok {
		return
	}
	if members, ok := h.Groups[groupName]; ok {
		delete(members, client)
		if len(members) == 0 {
			delete(h.Groups, groupName)
		}
		log.Printf("User %s left group %s", username, groupName)
	}
}

func (h *Hub) broadcastUserList() {
	userList := make([]string, 0, len(h.Clients))
	for username := range h.Clients {
		userList = append(userList, username)
	}

	message := &Message{
		Type:    "user_list",
		Content: userList,
	}

	for _, client := range h.Clients {
		select {
		case client.Send <- message:
		default:
			close(client.Send)
			delete(h.Clients, client.Username)
		}
	}
	log.Println("Broadcasted updated user list")
}
