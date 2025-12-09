package Server

func (s *Server) RegisterRoutes() {

	s.Router.GET("/", s.GetFirstSongs)
	s.Router.GET("/ping", s.Ping)
	s.Router.GET("/search", s.GetSearch)
}
