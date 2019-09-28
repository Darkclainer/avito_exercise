package main

func (s *Server) routes() {
	s.router.HandleFunc("/users/add", s.handleUserAdd()).Methods("POST")
}
