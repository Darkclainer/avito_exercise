package main

func (s *Server) routes() {
	s.router.HandleFunc("/users/add", s.handleAddUser()).Methods("POST")
	s.router.HandleFunc("/chats/add", s.handleAddChat()).Methods("POST")
	s.router.HandleFunc("/chats/get", s.handleGetUserChats()).Methods("POST")
	s.router.HandleFunc("/messages/add", s.handleAddMessage()).Methods("POST")
	s.router.HandleFunc("/messages/get", s.handleGetMessages()).Methods("POST")
}
