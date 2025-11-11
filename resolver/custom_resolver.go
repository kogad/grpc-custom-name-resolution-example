package resolver

import (
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc/resolver"
)

const scheme = "custom"

// customResolverBuilder implements resolver.Builder
type customResolverBuilder struct{}

// Build creates a new resolver for the given target
func (b *customResolverBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &customResolver{
		target: target,
		cc:     cc,
		addrs: []string{
			"localhost:8888",
			"localhost:9999",
		},
		currentIndex: 0,
		closeCh:      make(chan struct{}),
		updateCh:     make(chan []string, 1),
	}
	r.start()
	return r, nil
}

// Scheme returns the scheme supported by this resolver
func (b *customResolverBuilder) Scheme() string {
	return scheme
}

// customResolver implements resolver.Resolver
type customResolver struct {
	target       resolver.Target
	cc           resolver.ClientConn
	addrs        []string
	currentIndex int
	mu           sync.Mutex
	closeCh      chan struct{}
	updateCh     chan []string // Receives new address lists from external name resolution service
}

// start sends the initial address and starts listening for updates
func (r *customResolver) start() {
	// Send initial address (primary server: localhost:8888)
	r.updateAddresses()

	// Start monitoring for address updates from external name resolution service
	go func() {
		for {
			select {
			case newAddrs := <-r.updateCh:
				r.mu.Lock()
				r.addrs = newAddrs
				r.currentIndex = 0 // Reset to first address in new list
				r.mu.Unlock()
				r.updateAddresses()
				fmt.Printf("[Resolver] Address list updated from external service\n")
			case <-r.closeCh:
				return
			}
		}
	}()

	// Start mock name resolution service
	r.mockNameResolutionService()
}

// updateAddresses sends the current address to the ClientConn
func (r *customResolver) updateAddresses() {
	r.mu.Lock()
	addr := r.addrs[r.currentIndex]
	r.mu.Unlock()

	state := resolver.State{
		Addresses: []resolver.Address{
			{Addr: addr},
		},
	}
	r.cc.UpdateState(state)
	fmt.Printf("[Resolver] Updated address to: %s\n", addr)
}

// ResolveNow switches to the next available server when called by gRPC on connection failure
func (r *customResolver) ResolveNow(resolver.ResolveNowOptions) {
	r.mu.Lock()
	r.currentIndex = (r.currentIndex + 1) % len(r.addrs)
	r.mu.Unlock()
	r.updateAddresses()
}

func (r *customResolver) Close() {
	close(r.closeCh)
}

// mockNameResolutionService simulates an external name resolution service
// that periodically sends updated address lists
func (r *customResolver) mockNameResolutionService() {
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// Randomly shuffle the address order to simulate service discovery updates
				newAddrs := make([]string, len(r.addrs))
				r.mu.Lock()
				copy(newAddrs, r.addrs)
				r.mu.Unlock()

				// Shuffle: swap first and second elements
				newAddrs[0], newAddrs[1] = newAddrs[1], newAddrs[0]

				fmt.Printf("[NameResolutionService] Sending address update: %v\n", newAddrs)
				select {
				case r.updateCh <- newAddrs:
				case <-r.closeCh:
					return
				}
			case <-r.closeCh:
				return
			}
		}
	}()
}

func init() {
	resolver.Register(&customResolverBuilder{})
}
