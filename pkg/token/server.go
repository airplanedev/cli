package token

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/pkg/errors"
)

var (
	// DocsURL is the URL to redirect to after the token
	// has been sent on the channel.
	DocsURL = "https://github.com/airplanedev/cli"
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
	srv.server = &http.Server{
		Handler: srv,
		// Chrome, unlike FireFox/Safari, will preload a handful of connections as an
		// optimization, but unfortunately this means that we can't immediately shut down
		// the token server once we recieve the token, since there will be 1 or more pending
		// StateNew connections from Chrome. These get ignored after 5s during shutdown,
		// but that would cause `airplane login` to hang for 5s when logging in through Chrome.
		//
		// See: https://github.com/golang/go/issues/22682#issuecomment-343987847
		ReadHeaderTimeout: 500 * time.Millisecond,
	}
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
	fmt.Printf("received connection...\n")
	select {
	case <-r.Context().Done():
		fmt.Printf("closing from context...\n")
	case srv.tokens <- r.URL.Query().Get("token"):
		http.Redirect(w, r, DocsURL, http.StatusSeeOther)
		fmt.Printf("closing from token...\n")
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
	defer srv.wg.Wait()

	if err := srv.server.Shutdown(context.Background()); err != nil {
		fmt.Printf("server shutdown err: %+v", err)
		return errors.Wrap(err, "close server")
	}

	return nil
}
