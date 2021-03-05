package token

import (
	"context"
	"io"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
)

// Server implements a local token server.
//
// The server starts on a local random port and waits
// for a token request, when a token is received the
// server sends the token on the channel returned from `Tokens()`.
//
// It is important to configure the server with a shared
// context as it relies on it to shutdown in case a CLI
// login attempt is canceled.
//
//   srv, err := token.NewServer(ctx)
//
//   select {
//     case <-ctx.Done():
//       print("login canceled")
//     case token <- srv.Token():
//       verify(token)
//       save(token)
//   }
//
type Server struct {
	tokens chan string
	lstn   net.Listener
	ctx    context.Context
	wg     sync.WaitGroup
	server *http.Server
}

// NewServer returns a new server.
func NewServer(ctx context.Context) (*Server, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, errors.Wrap(err, "bind")
	}

	srv := &Server{
		tokens: make(chan string, 1),
		lstn:   l,
		ctx:    ctx,
	}
	srv.server = &http.Server{Handler: srv}
	srv.start()

	return srv, nil
}

// URL returns the server's URL.
func (srv *Server) URL() string {
	return "http://" + srv.lstn.Addr().String()
}

// Token returns the token channel.
func (srv *Server) Token() <-chan string {
	return srv.tokens
}

// ServeHTTP implementation.
func (srv *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
	case srv.tokens <- r.URL.Query().Get("token"):
		w.Header().Set("Content-Type", "text/html")
		io.WriteString(w, `<script>window.close()</script>`)
	}
}

// Start starts the server.
func (srv *Server) start() {
	srv.wg.Add(1)
	go func() {
		srv.server.Serve(srv.lstn)
		srv.wg.Done()
	}()
}

// Close closes the server.
func (srv *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	defer srv.wg.Wait()

	if err := srv.lstn.Close(); err != nil {
		srv.server.Shutdown(ctx)
		return errors.Wrap(err, "close listener")
	}

	if err := srv.server.Shutdown(ctx); err != nil {
		return errors.Wrap(err, "close server")
	}

	return nil
}
