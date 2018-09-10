// Copyright 2013 Julien Schmidt. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found
// at https://github.com/julienschmidt/httprouter/blob/master/LICENSE

package gateway

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/shisa-platform/core/context"
	"github.com/shisa-platform/core/httpx"
	"github.com/shisa-platform/core/service"
)

func printChildren(n *node, prefix string) {
	fmt.Printf(" %02d:%02d %s%s[%d] %q %#v %t %d \r\n", n.priority, n.maxParams, prefix, n.path, len(n.children), n.indices, n.endpoint, n.wildChild, n.nType)
	for l := len(n.path); l > 0; l-- {
		prefix += " "
	}
	for _, child := range n.children {
		printChildren(child, prefix)
	}
}

// Used as a workaround since we can't compare functions or their addressses
var fakeHandlerValue string

func fakeEndpoint(s string) *endpoint {
	return &endpoint{
		Endpoint: service.Endpoint{
			Route: s,
			Get: &service.Pipeline{
				Handlers: []httpx.Handler{
					func(context.Context, *httpx.Request) httpx.Response {
						fakeHandlerValue = s
						return nil
					},
				},
			},
		},
	}
}

type testRequests []struct {
	path        string
	nilEndpoint bool
	route       string
	ps          []httpx.PathParameter
}

func checkRequests(t *testing.T, tree *node, requests testRequests, unescapes ...bool) {
	t.Helper()
	for _, request := range requests {
		endpoint, ps, _, err := tree.getValue(request.path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if endpoint == nil {
			if !request.nilEndpoint {
				t.Errorf("endpoint mismatch for route %q: Expected non-nil endpoint", request.path)
			}
		} else if request.nilEndpoint {
			t.Errorf("endpoint mismatch for route %q: Expected nil endpoint", request.path)
		} else {
			endpoint.Get.Handlers[0](nil, nil)
			if fakeHandlerValue != request.route {
				t.Errorf("handle mismatch for route %q: Wrong handle (%s != %s)", request.path, fakeHandlerValue, request.route)
			}
		}

		if !reflect.DeepEqual(ps, request.ps) {
			t.Errorf("Params mismatch for route %q", request.path)
		}
	}
}

func checkPriorities(t *testing.T, n *node) uint32 {
	t.Helper()
	var prio uint32
	for i := range n.children {
		prio += checkPriorities(t, n.children[i])
	}

	if n.endpoint != nil {
		prio++
	}

	if n.priority != prio {
		t.Errorf(
			"priority mismatch for node %q: is %d, should be %d",
			n.path, n.priority, prio,
		)
	}

	return prio
}

func checkMaxParams(t *testing.T, n *node) uint8 {
	t.Helper()
	var maxParams uint8
	for i := range n.children {
		params := checkMaxParams(t, n.children[i])
		if params > maxParams {
			maxParams = params
		}
	}
	if n.nType > root && !n.wildChild {
		maxParams++
	}

	if n.maxParams != maxParams {
		t.Errorf(
			"maxParams mismatch for node %q: is %d, should be %d",
			n.path, n.maxParams, maxParams,
		)
	}

	return maxParams
}

func TestCountParams(t *testing.T) {
	if countParams("/path/:param1/static/*catch-all") != 2 {
		t.Fail()
	}
	if countParams(strings.Repeat("/:param", 256)) != 255 {
		t.Fail()
	}
}

func TestTreeAddAndGet(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/hi",
		"/contact",
		"/co",
		"/c",
		"/a",
		"/ab",
		"/doc/",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/α",
		"/β",
	}
	for _, route := range routes {
		if err := tree.addRoute(route, fakeEndpoint(route)); err != nil {
			t.Errorf("unexpected error adding route: %v", err)
		}
	}

	// printChildren(tree, "")

	checkRequests(t, tree, testRequests{
		{"/a", false, "/a", nil},
		{"/", true, "", nil},
		{"/hi", false, "/hi", nil},
		{"/contact", false, "/contact", nil},
		{"/co", false, "/co", nil},
		{"/con", true, "", nil},  // key mismatch
		{"/cona", true, "", nil}, // key mismatch
		{"/no", true, "", nil},   // no matching child
		{"/ab", false, "/ab", nil},
		{"/α", false, "/α", nil},
		{"/β", false, "/β", nil},
	})

	checkPriorities(t, tree)
	checkMaxParams(t, tree)
}

func TestTreeWildcard(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/",
		"/cmd/:tool/:sub",
		"/cmd/:tool/",
		"/src/*filepath",
		"/search/",
		"/search/:query",
		"/user_:name",
		"/user_:name/about",
		"/files/:dir",
		"/files/:dir/*filepath",
		"/doc/",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/info/:user/public",
		"/info/:user/project/:project",
	}
	for _, route := range routes {
		if err := tree.addRoute(route, fakeEndpoint(route)); err != nil {
			t.Errorf("unexpected error adding route: %v", err)
		}
	}

	// printChildren(tree, "")

	checkRequests(t, tree, testRequests{
		{"/", false, "/", nil},
		{"/cmd/test/", false, "/cmd/:tool/", []httpx.PathParameter{{Name: "tool", Value: "test"}}},
		{"/cmd/test", false, "/cmd/:tool/", []httpx.PathParameter{{Name: "tool", Value: "test"}}},
		{"/cmd/test/3", false, "/cmd/:tool/:sub", []httpx.PathParameter{{Name: "tool", Value: "test"}, {Name: "sub", Value: "3"}}},
		{"/src/", false, "/src/*filepath", []httpx.PathParameter{{Name: "filepath", Value: "/"}}},
		{"/src/some/file.png", false, "/src/*filepath", []httpx.PathParameter{{Name: "filepath", Value: "/some/file.png"}}},
		{"/search/", false, "/search/", nil},
		{"/search/someth!ng+in+ünìcodé", false, "/search/:query", []httpx.PathParameter{{Name: "query", Value: "someth!ng+in+ünìcodé"}}},
		{"/search/someth!ng+in+ünìcodé/", false, "/search/:query", []httpx.PathParameter{{Name: "query", Value: "someth!ng+in+ünìcodé"}}},
		{"/user_gopher", false, "/user_:name", []httpx.PathParameter{{Name: "name", Value: "gopher"}}},
		{"/user_gopher/about", false, "/user_:name/about", []httpx.PathParameter{{Name: "name", Value: "gopher"}}},
		{"/files/thingr/", false, "/files/:dir", []httpx.PathParameter{{Name: "dir", Value: "thingr"}}},
		{"/files/js/inc/framework.js", false, "/files/:dir/*filepath", []httpx.PathParameter{{Name: "dir", Value: "js"}, {Name: "filepath", Value: "/inc/framework.js"}}},
		{"/info/gordon/public", false, "/info/:user/public", []httpx.PathParameter{{Name: "user", Value: "gordon"}}},
		{"/info/gordon/project/go", false, "/info/:user/project/:project", []httpx.PathParameter{{Name: "user", Value: "gordon"}, {Name: "project", Value: "go"}}},
	})

	checkPriorities(t, tree)
	checkMaxParams(t, tree)
}

type testRoute struct {
	path     string
	conflict bool
}

func testRoutes(t *testing.T, routes []testRoute) {
	tree := &node{}

	for _, route := range routes {
		err := tree.addRoute(route.path, nil)

		if route.conflict {
			if err == nil {
				t.Errorf("expected error for conflicting route %q", route.path)
			}
		} else if err != nil {
			t.Errorf("unexpected error for route %q: %v", route.path, err)
		}
	}

	// printChildren(tree, "")
}

func TestTreeWildcardConflict(t *testing.T) {
	routes := []testRoute{
		{"/cmd/:tool/:sub", false},
		{"/cmd/vet", true},
		{"/src/*filepath", false},
		{"/src/*filepathx", true},
		{"/src/", true},
		{"/src1/", false},
		{"/src1/*filepath", true},
		{"/src2*filepath", true},
		{"/search/:query", false},
		{"/search/invalid", true},
		{"/user_:name", false},
		{"/user_x", true},
		{"/user_:name", false},
		{"/id:id", false},
		{"/id/:id", true},
	}
	testRoutes(t, routes)
}

func TestTreeChildConflict(t *testing.T) {
	routes := []testRoute{
		{"/cmd/vet", false},
		{"/cmd/:tool/:sub", true},
		{"/src/AUTHORS", false},
		{"/src/*filepath", true},
		{"/user_x", false},
		{"/user_:name", true},
		{"/id/:id", false},
		{"/id:id", true},
		{"/:id", true},
		{"/*filepath", true},
	}
	testRoutes(t, routes)
}

func TestTreeDupliatePath(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/",
		"/doc/",
		"/src/*filepath",
		"/search/:query",
		"/user_:name",
	}
	for _, route := range routes {
		if err := tree.addRoute(route, fakeEndpoint(route)); err != nil {
			t.Fatalf("unexpected error inserting route %q: %v", route, err)
		}

		// Add again
		if err := tree.addRoute(route, nil); err == nil {
			t.Fatalf("expected error while inserting duplicate route %q", route)
		}
	}

	// printChildren(tree, "")

	checkRequests(t, tree, testRequests{
		{"/", false, "/", nil},
		{"/doc/", false, "/doc/", nil},
		{"/src/some/file.png", false, "/src/*filepath", []httpx.PathParameter{{Name: "filepath", Value: "/some/file.png"}}},
		{"/search/someth!ng+in+ünìcodé", false, "/search/:query", []httpx.PathParameter{{Name: "query", Value: "someth!ng+in+ünìcodé"}}},
		{"/user_gopher", false, "/user_:name", []httpx.PathParameter{{Name: "name", Value: "gopher"}}},
	})
}

func TestEmptyWildcardName(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/user:",
		"/user:/",
		"/cmd/:/",
		"/src/*",
	}
	for _, route := range routes {
		if err := tree.addRoute(route, nil); err == nil {
			t.Fatalf("expected error while inserting route with empty wildcard name %q", route)
		}
	}
}

func TestTreeCatchAllConflict(t *testing.T) {
	routes := []testRoute{
		{"/src/*filepath/x", true},
		{"/src2/", false},
		{"/src2/*filepath/x", true},
	}
	testRoutes(t, routes)
}

func TestTreeCatchAllConflictRoot(t *testing.T) {
	routes := []testRoute{
		{"/", false},
		{"/*filepath", true},
	}
	testRoutes(t, routes)
}

func TestTreeDoubleWildcard(t *testing.T) {
	routes := [...]string{
		"/:foo:bar",
		"/:foo:bar/",
		"/:foo*bar",
	}

	for _, route := range routes {
		tree := &node{}

		if err := tree.addRoute(route, nil); err == nil {
			t.Fatalf("expected error for double wildcard on route %q", route)
		}
	}
}

func TestTreeTrailingSlashRedirect(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/hi",
		"/b/",
		"/search/:query",
		"/cmd/:tool/",
		"/src/*filepath",
		"/x",
		"/x/y",
		"/y/",
		"/y/z",
		"/0/:id",
		"/0/:id/1",
		"/1/:id/",
		"/1/:id/2",
		"/aa",
		"/a/",
		"/admin",
		"/admin/:category",
		"/admin/:category/:page",
		"/doc",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/no/a",
		"/no/b",
		"/api/hello/:name",
	}
	for _, route := range routes {
		if err := tree.addRoute(route, fakeEndpoint(route)); err != nil {
			t.Fatalf("unexpected error inserting route %q: %v", route, err)
		}
	}

	// printChildren(tree, "")

	tsrRoutes := [...]string{
		"/hi/",
		"/b",
		"/search/gopher/",
		"/cmd/vet",
		"/src",
		"/x/",
		"/y",
		"/0/go/",
		"/1/go",
		"/a",
		"/admin/",
		"/admin/config/",
		"/admin/config/permissions/",
		"/doc/",
	}
	for _, route := range tsrRoutes {
		endpoint, _, tsr, err := tree.getValue(route)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if endpoint == nil {
			t.Fatalf("expected non-nil endpoint for TSR route %q", route)
		}
		if !tsr {
			t.Errorf("expected TSR recommendation for route %q", route)
		}
	}

	noTsrRoutes := [...]string{
		"/",
		"/no",
		"/no/",
		"/_",
		"/_/",
		"/api/world/abc",
	}
	for _, route := range noTsrRoutes {
		endpoint, _, tsr, err := tree.getValue(route)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if endpoint != nil {
			t.Fatalf("expected nil endpoint for non-TSR route %q", route)
		}
		if tsr {
			t.Errorf("unexpected TSR recommendation for route %q", route)
		}
	}
}

func TestTreeRootTrailingSlashRedirect(t *testing.T) {
	tree := &node{}

	if err := tree.addRoute("/:test", fakeEndpoint("/:test")); err != nil {
		t.Fatalf("unexpected error inserting test route: %v", err)
	}

	endpoint, _, tsr, err := tree.getValue("/")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if endpoint != nil {
		t.Fatalf("expected nil endpoint")
	}
	if tsr {
		t.Errorf("unexpected TSR recommendation")
	}
}

func TestTreeFindCaseInsensitivePath(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/hi",
		"/b/",
		"/ABC/",
		"/search/:query",
		"/cmd/:tool/",
		"/src/*filepath",
		"/x",
		"/x/y",
		"/y/",
		"/y/z",
		"/0/:id",
		"/0/:id/1",
		"/1/:id/",
		"/1/:id/2",
		"/aa",
		"/a/",
		"/doc",
		"/doc/go_faq.html",
		"/doc/go1.html",
		"/doc/go/away",
		"/no/a",
		"/no/b",
	}

	for _, route := range routes {
		if err := tree.addRoute(route, fakeEndpoint(route)); err != nil {
			t.Fatalf("unexpected error inserting route %q: %v", route, err)
		}
	}

	// Check out == in for all registered routes
	// With fixTrailingSlash = true
	for _, route := range routes {
		out, found, err := tree.findCaseInsensitivePath(route, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !found {
			t.Errorf("Route %q not found!", route)
		} else if string(out) != route {
			t.Errorf("Wrong result for route %q: %s", route, string(out))
		}
	}
	// With fixTrailingSlash = false
	for _, route := range routes {
		out, found, err := tree.findCaseInsensitivePath(route, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !found {
			t.Errorf("Route %q not found!", route)
		} else if string(out) != route {
			t.Errorf("Wrong result for route %q: %s", route, string(out))
		}
	}

	tests := []struct {
		in    string
		out   string
		found bool
		slash bool
	}{
		{"/HI", "/hi", true, false},
		{"/HI/", "/hi", true, true},
		{"/B", "/b/", true, true},
		{"/B/", "/b/", true, false},
		{"/abc", "/ABC/", true, true},
		{"/abc/", "/ABC/", true, false},
		{"/aBc", "/ABC/", true, true},
		{"/aBc/", "/ABC/", true, false},
		{"/abC", "/ABC/", true, true},
		{"/abC/", "/ABC/", true, false},
		{"/SEARCH/QUERY", "/search/QUERY", true, false},
		{"/SEARCH/QUERY/", "/search/QUERY", true, true},
		{"/CMD/TOOL/", "/cmd/TOOL/", true, false},
		{"/CMD/TOOL", "/cmd/TOOL/", true, true},
		{"/SRC/FILE/PATH", "/src/FILE/PATH", true, false},
		{"/x/Y", "/x/y", true, false},
		{"/x/Y/", "/x/y", true, true},
		{"/X/y", "/x/y", true, false},
		{"/X/y/", "/x/y", true, true},
		{"/X/Y", "/x/y", true, false},
		{"/X/Y/", "/x/y", true, true},
		{"/Y/", "/y/", true, false},
		{"/Y", "/y/", true, true},
		{"/Y/z", "/y/z", true, false},
		{"/Y/z/", "/y/z", true, true},
		{"/Y/Z", "/y/z", true, false},
		{"/Y/Z/", "/y/z", true, true},
		{"/y/Z", "/y/z", true, false},
		{"/y/Z/", "/y/z", true, true},
		{"/Aa", "/aa", true, false},
		{"/Aa/", "/aa", true, true},
		{"/AA", "/aa", true, false},
		{"/AA/", "/aa", true, true},
		{"/aA", "/aa", true, false},
		{"/aA/", "/aa", true, true},
		{"/A/", "/a/", true, false},
		{"/A", "/a/", true, true},
		{"/DOC", "/doc", true, false},
		{"/DOC/", "/doc", true, true},
		{"/NO", "", false, true},
		{"/DOC/GO", "", false, true},
	}
	// With fixTrailingSlash = true
	for _, test := range tests {
		out, found, err := tree.findCaseInsensitivePath(test.in, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if found != test.found || (found && (string(out) != test.out)) {
			t.Errorf("Wrong result for %q: got %s, %t; want %s, %t",
				test.in, string(out), found, test.out, test.found)
			return
		}
	}
	// With fixTrailingSlash = false
	for _, test := range tests {
		out, found, err := tree.findCaseInsensitivePath(test.in, false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if test.slash {
			if found { // test needs a trailingSlash fix. It must not be found!
				t.Errorf("Found without fixTrailingSlash: %s; got %s", test.in, string(out))
			}
		} else {
			if found != test.found || (found && (string(out) != test.out)) {
				t.Errorf("Wrong result for %q: got %s, %t; want %s, %t",
					test.in, string(out), found, test.out, test.found)
				return
			}
		}
	}
}

func TestTreeInvalidNodeType(t *testing.T) {
	tree := &node{}
	if err := tree.addRoute("/", fakeEndpoint("/")); err != nil {
		t.Fatalf("unexpected error adding fixture: %v", err)
	}
	if err := tree.addRoute("/:page", fakeEndpoint("/:page")); err != nil {
		t.Fatalf("unexpected error adding fixture: %v", err)
	}

	// set invalid node type
	tree.children[0].nType = 42

	// normal lookup
	if _, _, _, err := tree.getValue("/test"); err == nil {
		t.Fatalf("expected error")
	}

	// case-insensitive lookup
	if _, _, err := tree.findCaseInsensitivePath("/test", true); err == nil {
		t.Fatalf("expected error")
	}
}

func TestTreeMultipleNamedParameters(t *testing.T) {
	tree := &node{}

	routes := [...]string{
		"/thing/:foo/:bar",
		"/thing/:foo",
		"/thing",
	}

	for _, route := range routes {
		if err := tree.addRoute(route, fakeEndpoint(route)); err != nil {
			t.Fatalf("unexpected error inserting route %q: %v", route, err)
		}
	}

	// printChildren(tree, "")

	tests := [...]string{
		"/thing/this/",
		"/thing/",
	}
	for _, route := range tests {
		endpoint, _, tsr, err := tree.getValue(route)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if endpoint == nil {
			t.Fatalf("expected non-nil endpoint for route %q", route)
		}
		if !tsr {
			t.Errorf("expected TSR recommendation for route %q", route)
		}
	}
}
