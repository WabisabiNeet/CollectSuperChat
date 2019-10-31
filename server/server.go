package server

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Server struct.
type Server struct {
	addWatchingLiveStream func(vid string)
}

// NewServer return new Server struct.
func NewServer(addLiveStreamFunc func(string)) *Server {
	return &Server{
		addWatchingLiveStream: addLiveStreamFunc,
	}
}

// Serve listen request
func (s *Server) Serve() {
	listener, err := net.Listen("tcp", "127.0.0.1:10080")
	if err != nil {
		panic(err)
	}

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)
	go func() {
		<-sig
		listener.Close()
	}()

	mux := http.NewServeMux()
	mux.HandleFunc("/watch", s.handleWatch)
	ch := make(chan error)
	go func() {
		ch <- http.Serve(listener, mux)
	}()
	fmt.Println("Server started at ", listener.Addr())
	// fmt.Println(<-ch)
}

func (s *Server) handleWatch(w http.ResponseWriter, r *http.Request) {

	// HTTPメソッドをチェック（POSTのみ許可）
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed) // 405
		w.Write([]byte("post only method."))
		return
	}

	r.ParseForm()
	params := r.PostForm
	vid := params["vid"]

	if len(vid) == 0 {
		w.WriteHeader(http.StatusBadRequest) // 400
		w.Write([]byte("Invalid parameter."))
		return
	}

	s.addWatchingLiveStream(vid[0])
	w.Write([]byte(fmt.Sprintf("watch start: %v", vid[0])))
	return
}
