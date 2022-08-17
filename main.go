package main

import (
	"context"
	"fmt"
	"github.com/perbu/dns-reliability/config"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

type res struct {
	server string
	time   time.Time
	dur    time.Duration
	err    error
}

type resCh chan res

func main() {
	err := realMain()
	if err != nil {
		log.Fatalln("error from realMain: ", err)
	}
}

func realMain() error {
	c, err := config.ParseConfigFile("config.yaml")
	if err != nil {
		return fmt.Errorf("config.ParseConfigFile: %w", err)
	}
	fmt.Printf("Poll interval is %v\n", c.Interval)
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	// ctx, cancel := context.WithTimeout(context.Background(), time.Second*7)
	defer cancel()
	err = runMonitor(ctx, c)
	return err
}

func runMonitor(ctx context.Context, c config.Config) error {
	ch := make(resCh, 100)
	wg := sync.WaitGroup{}
	for _, provider := range c.DNS {
		for _, server := range provider.Servers {
			if server.Ipv4 != "" {
				wg.Add(1)
				go func(s config.Server) {
					defer wg.Done()
					fmt.Printf("Monitoring %s [%s]: query %s\n", s.Name, s.Ipv4, s.Query)
					monitorServer(ctx, ch, s, c.Interval)
				}(server)
			}
			if server.Ipv6 != "" {
				// ignore ipv6 for now
				// go monitorServer(server, ch)
			}
		}
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		processResults(ctx, ch)
	}()
	wg.Wait()
	close(ch)
	return nil
}

func processResults(ctx context.Context, ch resCh) {
	type mres struct {
		latency time.Duration
		ts      time.Time
		err     error
	}
	type curr struct {
		ts  time.Time
		err error
	}
	// 	var data map[string][]mres
	data := make(map[string][]mres)
loop:
	for {
		select {
		case <-ctx.Done():
			fmt.Print("\n")
			break loop
		case r := <-ch:
			lst, ok := data[r.server]
			if !ok {
				lst = []mres{}
				data[r.server] = lst
			}
			data[r.server] = append(data[r.server], mres{
				latency: r.dur,
				ts:      r.time,
				err:     r.err,
			})
			fmt.Print(".")
		}
	}
	fmt.Println("======= Interrupt =======")
	for k, v := range data {
		failures := 0
		successes := 0
		errs := make([]curr, 0)
		for _, r := range v {
			if r.err != nil {
				failures++
				errs = append(errs, curr{err: r.err, ts: r.ts})
			} else {
				successes++
			}
		}
		fmt.Printf("%s: sucessses: %d, failures: %d\n", k, successes, failures)
		for _, r := range errs {
			fmt.Printf("   - %v: %s\n", r.ts, r.err)
		}
	}
}

func monitorServer(ctx context.Context, ch resCh, entry config.Server, interval time.Duration) {
	resolver, err := makeResolver(entry.Ipv4)
	if err != nil {
		// todo
		panic(err)
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for ctx.Err() == nil {
		<-ticker.C
		start := time.Now()
		_, err := resolver.LookupHost(ctx, entry.Query)
		dur := time.Since(start)
		ch <- res{
			server: entry.Name,
			time:   start,
			dur:    dur,
			err:    err,
		}
	}
}

func makeResolver(addr string) (*net.Resolver, error) {
	r := &net.Resolver{
		PreferGo: true,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: time.Second,
			}
			//log.Printf("dialing %s with addr %s:53 network: %s\n", address, addr, network)
			return d.DialContext(ctx, network, addr+":53")
		},
	}
	return r, nil
}
