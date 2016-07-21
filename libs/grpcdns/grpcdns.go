package grpcdns

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/sprucehealth/backend/libs/golog"
	"google.golang.org/grpc/naming"
)

type lookuper interface {
	lookup() ([]string, error)
}

type srvLookuper struct {
	service, proto, domain string
}

func (l *srvLookuper) lookup() ([]string, error) {
	_, srvs, err := net.LookupSRV(l.service, l.proto, l.domain)
	if err != nil {
		return nil, err
	}
	addrs := make([]string, len(srvs))
	for i, s := range srvs {
		addrs[i] = fmt.Sprintf("%s:%d", strings.TrimRight(s.Target, "."), s.Port)
	}
	return addrs, nil
}

type hostPortLookuper struct {
	host string
	port int
}

func (l *hostPortLookuper) lookup() ([]string, error) {
	ips, err := net.LookupHost(l.host)
	if err != nil {
		return nil, err
	}
	addrs := make([]string, len(ips))
	for i, ip := range ips {
		addrs[i] = fmt.Sprintf("%s:%d", ip, l.port)
	}
	return addrs, nil
}

type resolver struct {
	interval time.Duration
}

type watcher struct {
	target   string
	lookuper lookuper
	stopCh   chan bool
	interval time.Duration
	updateCh chan []*naming.Update
	addr     map[string]struct{} // currently known set of hosts
}

// Resolver accepts either a host:port or just a host in which case it expects the
// host to be a SRV lookup of the form _service._protocol.domain.
func Resolver(interval time.Duration) naming.Resolver {
	return &resolver{interval: interval}
}

// Resolve creates a Watcher for target.
func (r *resolver) Resolve(target string) (naming.Watcher, error) {
	l, err := lookuperFromTarget(target)
	if err != nil {
		return nil, err
	}
	w := &watcher{
		lookuper: l,
		target:   target,
		stopCh:   make(chan bool),
		interval: r.interval,
		updateCh: make(chan []*naming.Update),
		addr:     make(map[string]struct{}),
	}
	go w.loop()
	return w, nil
}

// Next blocks until an update or error happens. It may return one or more
// updates. The first call should get the full set of the results. It should
// return an error if and only if Watcher cannot recover.
func (w *watcher) Next() ([]*naming.Update, error) {
	return <-w.updateCh, nil
}

// Close closes the Watcher.
func (w *watcher) Close() {
	close(w.stopCh)
}

func (w *watcher) loop() {
	tick := time.NewTicker(w.interval)
	defer tick.Stop()
	for {
		if updates, err := w.update(); err != nil {
			golog.Errorf(err.Error())
		} else if len(updates) != 0 {
			w.updateCh <- updates
		}
		select {
		case <-w.stopCh:
			return
		case <-tick.C:
		}
	}
}

func (w *watcher) update() ([]*naming.Update, error) {
	addrs, err := w.lookuper.lookup()
	if err != nil {
		return nil, fmt.Errorf("grpcdns: failed to lookup '%s': %s", w.target, err)
	}

	// Don't do anything for an empty set of hosts as it's better to try to talk
	// to dead hosts then to accidently lose the entire host list.
	if len(addrs) == 0 {
		return nil, fmt.Errorf("grpcdns: empty host list for '%s'", w.target)
	}

	var updates []*naming.Update
	newSet := make(map[string]struct{}, len(addrs))
	for _, a := range addrs {
		newSet[a] = struct{}{}
		if _, ok := w.addr[a]; ok {
			delete(w.addr, a)
		} else {
			updates = append(updates, &naming.Update{
				Op:   naming.Add,
				Addr: a,
			})
		}
	}
	// anything left no longer exists
	for a := range w.addr {
		updates = append(updates, &naming.Update{
			Op:   naming.Delete,
			Addr: a,
		})
	}
	w.addr = newSet

	return updates, nil
}

func lookuperFromTarget(target string) (lookuper, error) {
	if ix := strings.Index(target, ":"); ix >= 0 {
		host := target[:ix]
		port, err := strconv.Atoi(target[ix+1:])
		if err != nil {
			return nil, fmt.Errorf("grpcdns: failed to parse port '%s'", target[ix+1:])
		}
		return &hostPortLookuper{
			host: host,
			port: port,
		}, nil
	}

	parts := strings.SplitN(target, ".", 3)
	if len(parts) != 3 {
		return nil, errors.New("grpcdns: srv hostnames must be of the form _service._protocol.domain")
	}
	if !strings.HasPrefix(parts[0], "_") {
		return nil, fmt.Errorf("grpcdns: missing _ on first part of srv target '%s'", target)
	}
	if !strings.HasPrefix(parts[1], "_") {
		return nil, fmt.Errorf("grpcdns: missing _ on first part of srv target '%s'", target)
	}
	return &srvLookuper{
		service: parts[0][1:],
		proto:   parts[1][1:],
		domain:  parts[2],
	}, nil
}
