package server

import (
	"fmt"
	"net/http"
)

func main() {

	// POSTのみ許可.
	http.HandleFunc("/watch", handleWatch)

	// 8080ポートで起動
	http.ListenAndServe(":10080", nil)
}

// Server struct.
type Server struct {
}

// Serve listen request
func (s *Server) Serve() {

	http.HandleFunc("/watch", handleWatch)

	// 8080ポートで起動
	http.ListenAndServe(":10080", nil)

}

func handleWatch(w http.ResponseWriter, r *http.Request) {

	// HTTPメソッドをチェック（POSTのみ許可）
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed) // 405
		w.Write([]byte("post only method."))
		return
	}

	r.ParseForm()
	params := r.PostForm
	channel := params["channel"]

	w.Write([]byte(fmt.Sprintf("watching start:%v", channel)))
}
