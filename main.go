package main // import "github.com/dedelala/patchy"

// #cgo pkg-config: jack
// #include <errno.h>
// #include <jack/jack.h>
// #include <jack/types.h>
//
// static char* at(int i, char** ss) {
// 	return ss[i];
// }
//
// static jack_client_t* client(jack_status_t* s) {
// 	return jack_client_open("patchy", JackNullOption, s);
// }
import "C"
import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

type client *C.jack_client_t
type patch map[string]map[string]bool

func get(c client) patch {
	srcs := smap(C.jack_get_ports(c, nil, nil, C.JackPortIsOutput))
	m := patch{}
	for src := range srcs {
		h := C.jack_port_by_name(c, C.CString(src))
		dsts := smap(C.jack_port_get_all_connections(c, h))
		if len(dsts) > 0 {
			m[src] = dsts
		}
	}
	return m
}

func load(path string) (patch, error) {
	var r io.Reader
	p := patch{}
	switch {
	case path == "" || path == "-":
		r = os.Stdin
	default:
		f, err := os.Open(path)
		if err != nil {
			return p, err
		}
		defer f.Close()
		r = f
	}
	err := json.NewDecoder(r).Decode(&p)
	return p, err
}

func (p patch) disco(c client) error {
	q := get(c)
	for s, ds := range q {
		for d := range ds {
			if p[s] == nil || !p[s][d] {
				i := C.jack_disconnect(c, C.CString(s), C.CString(d))
				if i != 0 {
					return fmt.Errorf("some kind of error disconnecting %q from %q", s, d)
				}
			}
		}
	}
	return nil
}

func (p patch) recall(c client) error {
	ports := smap(C.jack_get_ports(c, nil, nil, 0))
	for s, ds := range p {
		if !ports[s] {
			return fmt.Errorf("no port: %q", s)
		}
		for d := range ds {
			if !ports[d] {
				return fmt.Errorf("no port: %q", d)
			}
		}
	}
	if err := p.disco(c); err != nil {
		return err
	}
	for s, ds := range p {
		for d := range ds {
			i := C.jack_connect(c, C.CString(s), C.CString(d))
			if i != 0 && i != C.EEXIST {
				return fmt.Errorf("some kind of error connecting %q to %q", s, d)
			}
		}
	}
	return nil
}

func (p patch) store(path string) error {
	var w io.Writer
	switch {
	case path == "" || path == "-":
		w = os.Stdout
	default:
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		defer f.Close()
		w = f
	}
	return json.NewEncoder(w).Encode(p)
}

func smap(cs **C.char) map[string]bool {
	m := map[string]bool{}
	if cs == nil {
		return m
	}
	for i := C.int(0); true; i++ {
		s := C.GoString(C.at(i, cs))
		if s == "" {
			break
		}
		m[s] = true
	}
	return m
}

func main() {
	log.SetFlags(0)
	flag.Parse()

	sto := func(c client, a string) error {
		return get(c).store(a)
	}
	rcl := func(c client, a string) error {
		p, err := load(a)
		if err != nil {
			return err
		}
		return p.recall(c)
	}
	fs := map[string]func(client, string) error{
		"s": sto, "sto": sto, "store": sto,
		"r": rcl, "rcl": rcl, "recall": rcl,
	}
	f, ok := fs[flag.Arg(0)]
	if !ok {
		log.Printf("usage: %s [store|recall] [file|-]", os.Args[0])
		os.Exit(2)
	}

	var s C.jack_status_t
	c := C.client(&s)
	if s != 0 {
		log.Fatalf("jack client failed: %b", s)
	}

	if err := f(c, flag.Arg(1)); err != nil {
		log.Fatal(err)
	}
}
