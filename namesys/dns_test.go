package namesys

import (
	"fmt"
	"testing"

	opts "github.com/dms3-fs/go-dms3-fs/namesys/opts"
)

type mockDNS struct {
	entries map[string][]string
}

func (m *mockDNS) lookupTXT(name string) (txt []string, err error) {
	txt, ok := m.entries[name]
	if !ok {
		return nil, fmt.Errorf("no TXT entry for %s", name)
	}
	return txt, nil
}

func TestDnsEntryParsing(t *testing.T) {

	goodEntries := []string{
		"QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD",
		"dnslink=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD",
		"dnslink=/dms3ns/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD",
		"dnslink=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/foo",
		"dnslink=/dms3ns/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/bar",
		"dnslink=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/foo/bar/baz",
		"dnslink=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/foo/bar/baz/",
		"dnslink=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD",
	}

	badEntries := []string{
		"QmYhE8xgFCjGcz6PHgnvJz5NOTCORRECT",
		"quux=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD",
		"dnslink=",
		"dnslink=/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/foo",
		"dnslink=dms3ns/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/bar",
	}

	for _, e := range goodEntries {
		_, err := parseEntry(e)
		if err != nil {
			t.Log("expected entry to parse correctly!")
			t.Log(e)
			t.Fatal(err)
		}
	}

	for _, e := range badEntries {
		_, err := parseEntry(e)
		if err == nil {
			t.Log("expected entry parse to fail!")
			t.Fatal(err)
		}
	}
}

func newMockDNS() *mockDNS {
	return &mockDNS{
		entries: map[string][]string{
			"multihash.example.com": []string{
				"dnslink=QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD",
			},
			"dms3fs.example.com": []string{
				"dnslink=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD",
			},
			"_dnslink.ddms3fs.example.com": []string{
				"dnslink=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD",
			},
			"dns1.example.com": []string{
				"dnslink=/dms3ns/dms3fs.example.com",
			},
			"dns2.example.com": []string{
				"dnslink=/dms3ns/dns1.example.com",
			},
			"multi.example.com": []string{
				"some stuff",
				"dnslink=/dms3ns/dns1.example.com",
				"masked dnslink=/dms3ns/example.invalid",
			},
			"equals.example.com": []string{
				"dnslink=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/=equals",
			},
			"loop1.example.com": []string{
				"dnslink=/dms3ns/loop2.example.com",
			},
			"loop2.example.com": []string{
				"dnslink=/dms3ns/loop1.example.com",
			},
			"_dnslink.dloop1.example.com": []string{
				"dnslink=/dms3ns/loop2.example.com",
			},
			"_dnslink.dloop2.example.com": []string{
				"dnslink=/dms3ns/loop1.example.com",
			},
			"bad.example.com": []string{
				"dnslink=",
			},
			"withsegment.example.com": []string{
				"dnslink=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/sub/segment",
			},
			"withrecsegment.example.com": []string{
				"dnslink=/dms3ns/withsegment.example.com/subsub",
			},
			"withtrailing.example.com": []string{
				"dnslink=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/sub/",
			},
			"withtrailingrec.example.com": []string{
				"dnslink=/dms3ns/withtrailing.example.com/segment/",
			},
			"double.example.com": []string{
				"dnslink=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD",
			},
			"_dnslink.double.example.com": []string{
				"dnslink=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD",
			},
			"double.conflict.com": []string{
				"dnslink=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD",
			},
			"_dnslink.conflict.example.com": []string{
				"dnslink=/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjE",
			},
		},
	}
}

func TestDNSResolution(t *testing.T) {
	mock := newMockDNS()
	r := &DNSResolver{lookupTXT: mock.lookupTXT}
	testResolution(t, r, "multihash.example.com", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD", nil)
	testResolution(t, r, "dms3fs.example.com", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD", nil)
	testResolution(t, r, "ddms3fs.example.com", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD", nil)
	testResolution(t, r, "dns1.example.com", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD", nil)
	testResolution(t, r, "dns1.example.com", 1, "/dms3ns/dms3fs.example.com", ErrResolveRecursion)
	testResolution(t, r, "dns2.example.com", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD", nil)
	testResolution(t, r, "dns2.example.com", 1, "/dms3ns/dns1.example.com", ErrResolveRecursion)
	testResolution(t, r, "dns2.example.com", 2, "/dms3ns/dms3fs.example.com", ErrResolveRecursion)
	testResolution(t, r, "multi.example.com", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD", nil)
	testResolution(t, r, "multi.example.com", 1, "/dms3ns/dns1.example.com", ErrResolveRecursion)
	testResolution(t, r, "multi.example.com", 2, "/dms3ns/dms3fs.example.com", ErrResolveRecursion)
	testResolution(t, r, "equals.example.com", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/=equals", nil)
	testResolution(t, r, "loop1.example.com", 1, "/dms3ns/loop2.example.com", ErrResolveRecursion)
	testResolution(t, r, "loop1.example.com", 2, "/dms3ns/loop1.example.com", ErrResolveRecursion)
	testResolution(t, r, "loop1.example.com", 3, "/dms3ns/loop2.example.com", ErrResolveRecursion)
	testResolution(t, r, "loop1.example.com", opts.DefaultDepthLimit, "/dms3ns/loop1.example.com", ErrResolveRecursion)
	testResolution(t, r, "dloop1.example.com", 1, "/dms3ns/loop2.example.com", ErrResolveRecursion)
	testResolution(t, r, "dloop1.example.com", 2, "/dms3ns/loop1.example.com", ErrResolveRecursion)
	testResolution(t, r, "dloop1.example.com", 3, "/dms3ns/loop2.example.com", ErrResolveRecursion)
	testResolution(t, r, "dloop1.example.com", opts.DefaultDepthLimit, "/dms3ns/loop1.example.com", ErrResolveRecursion)
	testResolution(t, r, "bad.example.com", opts.DefaultDepthLimit, "", ErrResolveFailed)
	testResolution(t, r, "withsegment.example.com", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/sub/segment", nil)
	testResolution(t, r, "withrecsegment.example.com", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/sub/segment/subsub", nil)
	testResolution(t, r, "withsegment.example.com/test1", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/sub/segment/test1", nil)
	testResolution(t, r, "withrecsegment.example.com/test2", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/sub/segment/subsub/test2", nil)
	testResolution(t, r, "withrecsegment.example.com/test3/", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/sub/segment/subsub/test3/", nil)
	testResolution(t, r, "withtrailingrec.example.com", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD/sub/segment/", nil)
	testResolution(t, r, "double.example.com", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjD", nil)
	testResolution(t, r, "conflict.example.com", opts.DefaultDepthLimit, "/dms3fs/QmY3hE8xgFCjGcz6PHgnvJz5HZi1BaKRfPkn1ghZUcYMjE", nil)
}
