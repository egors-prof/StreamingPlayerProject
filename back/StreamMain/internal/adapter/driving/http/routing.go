package http

func (s *Server) endpoints() {
	streamG := s.router.Group("/stream")
	streamG.GET("/ws", s.WebsocketHandler)
}
