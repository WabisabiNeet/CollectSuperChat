package arcive

import (
	"net/http"

	"github.com/ant0ine/go-json-rest/rest"
)

type ArchiveServer struct {
	CollectArchiveChat func(vid string)
}

func NewArchiveServer(f func(string, bool)) *ArchiveServer {
	return &ArchiveServer{
		CollectChat: f,
	}
}

func (s *ArchiveServer) Serve() {
	api := rest.NewApi()
	api.Use(rest.DefaultDevStack...)
	router, err := rest.MakeRouter(
		rest.Put("/archive/:vid", s.watchArchive),
	)
	if err != nil {
		// log.Fatal(err)
	}
	api.SetApp(router)
	http.ListenAndServe(":8080", api.MakeHandler())
}

func (s *ArchiveServer) watchArchive(w rest.ResponseWriter, r *rest.Request) {
	vid := r.PathParam("vid")
	if vid == "" {
		w.WriteHeader(http.StatusBadRequest)
	}

	s.CollectArchiveChat(vid)
}
