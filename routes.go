package main

func (s *Server) routes() {
	s.router.HandleFunc("/users/add", s.handleAddUser()).Methods("POST")
	s.router.HandleFunc("/chats/add", s.handleAddChat()).Methods("POST")
}
