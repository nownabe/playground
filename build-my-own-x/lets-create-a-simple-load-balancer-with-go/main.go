package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type contextKey int

const (
	keyAttempts contextKey = iota
	keyRetries
)

type backend struct {
	url   *url.URL
	alive bool
	mux   sync.RWMutex
	proxy *httputil.ReverseProxy
}

func newBackend(host string) (*backend, error) {
	u, err := url.Parse(host)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %s: %w", host, err)
	}

	return &backend{
		url:   u,
		alive: true,
		proxy: httputil.NewSingleHostReverseProxy(u),
	}, nil
}

func (b *backend) setAlive(alive bool) {
	b.mux.Lock()
	defer b.mux.Unlock()
	b.alive = alive
}

func (b *backend) isAlive() bool {
	b.mux.RLock()
	defer b.mux.RUnlock()
	return b.alive
}

func (b *backend) healthcheck() {
	timeout := 2 * time.Second

	conn, err := net.DialTimeout("tcp", b.url.Host, timeout)
	if err != nil {
		b.setAlive(false)
		log.Println("Site unreachable, error: ", err)
		return
	}

	_ = conn.Close()
	b.setAlive(true)
}

func (b *backend) setErrorHandler(fn func(http.ResponseWriter, *http.Request, error)) {
	b.proxy.ErrorHandler = fn
}

type serverPool struct {
	backends []*backend
	current  uint64
}

func (s *serverPool) addBackend(b *backend) {
	h := func(w http.ResponseWriter, r *http.Request, e error) {
		log.Printf("[%s] %s\n", r.RemoteAddr, e.Error())
		retries := getRetryFromContext(r)
		if retries < 3 {
			time.Sleep(10 * time.Millisecond)
			ctx := context.WithValue(r.Context(), keyRetries, retries+1)
			b.proxy.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		attempts := getAttemptsFromContext(r)
		log.Printf("%s(%s) Attempting retry %d\n", r.RemoteAddr, r.URL.Path, attempts)
		ctx := context.WithValue(r.Context(), keyAttempts, attempts+1)
		s.serve(w, r.WithContext(ctx))
	}

	b.setErrorHandler(h)

	s.backends = append(s.backends, b)
}

func (s *serverPool) nextIndex() int {
	return int(atomic.AddUint64(&s.current, uint64(1)) % uint64(len(s.backends)))
}

func (s *serverPool) getNextPeer() *backend {
	next := s.nextIndex()
	l := len(s.backends) + next

	for i := next; i < l; i++ {
		idx := i % len(s.backends)
		if s.backends[idx].isAlive() {
			if i != next {
				atomic.StoreUint64(&s.current, uint64(idx))
			}
			return s.backends[idx]
		}
	}

	return nil
}

func (s *serverPool) healthcheck() {
	for _, b := range s.backends {
		b.healthcheck()
	}
}

func (s *serverPool) startHealthcheck() {
	t := time.NewTicker(time.Second * 30)

	for range t.C {
		log.Println("Starting health check...")
		s.healthcheck()
		log.Println("Health check completed")
	}
}

func (s *serverPool) serve(w http.ResponseWriter, r *http.Request) {
	attempts := getAttemptsFromContext(r)
	if attempts > 3 {
		log.Printf("%s(%s) Max attempts reached, terminating\n", r.RemoteAddr, r.URL.Path)
		http.Error(w, "service not available", http.StatusServiceUnavailable)
		return
	}

	peer := s.getNextPeer()
	if peer != nil {
		peer.proxy.ServeHTTP(w, r)
		return
	}
	http.Error(w, "Service not available", http.StatusServiceUnavailable)
}

func main() {
	var serversList string
	var port int

	flag.StringVar(&serversList, "backends", "", "Load balanced backends, use comma as separator")
	flag.IntVar(&port, "port", 3030, "Port to listen")
	flag.Parse()

	if len(serversList) == 0 {
		log.Fatal("please provide one or more backends to load balance")
	}

	pool := &serverPool{}

	servers := strings.Split(serversList, ",")
	for _, server := range servers {
		b, err := newBackend(server)
		if err != nil {
			log.Fatal(err)
		}
		pool.addBackend(b)
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: http.HandlerFunc(pool.serve),
	}

	go pool.startHealthcheck()

	log.Printf("Load balancer started at:%d", port)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func getRetryFromContext(r *http.Request) int {
	if retry, ok := r.Context().Value(keyRetries).(int); ok {
		return retry
	}
	return 0
}

func getAttemptsFromContext(r *http.Request) int {
	if attempts, ok := r.Context().Value(keyAttempts).(int); ok {
		return attempts
	}
	return 0
}
