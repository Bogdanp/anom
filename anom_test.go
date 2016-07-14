package anom

import (
	"flag"
	"log"
	"os"
	"testing"

	"golang.org/x/net/context"

	"google.golang.org/appengine/aetest"
	"google.golang.org/appengine/datastore"
)

var ctx context.Context
var clark *User

const knownID = "clark"

type User struct {
	Meta
	Username string
}

func (u *User) GetMeta() *Meta {
	return &u.Meta
}

type Post struct {
	Meta
	Content string
}

func (p *Post) GetMeta() *Meta {
	return &p.Meta
}

func TestGet(t *testing.T) {
	u := &User{}
	if err := Get(ctx, u, WithStringID(ctx, knownID)); err != nil {
		t.Fatal(err)
	}

	if *u.Key != *clark.Key {
		t.Fatalf("user isn't Clark")
	}
}

func TestGetErrors(t *testing.T) {
	u := &User{}
	tests := []struct {
		result   error
		expected error
	}{
		{Get(ctx, u), ErrMissingKey},
		{Get(ctx, u, WithStringID(ctx, "idontexist")), datastore.ErrNoSuchEntity},
		{Get(ctx, u, WithIntID(ctx, 25)), datastore.ErrNoSuchEntity},
	}

	for i, test := range tests {
		if test.result != test.expected {
			t.Errorf("expected %v but got %v for case %d", test.expected, test.result, i)
		}
	}
}

func TestPut(t *testing.T) {
	u := &User{Username: "Jim"}
	if err := Put(ctx, u); err != nil {
		t.Fatal(err)
	}

	if u.Meta.Key == nil {
		t.Fatal("key missing after put")
	}

	u2 := &User{}
	if err := Get(ctx, u2, WithKey(u.Key)); err != nil {
		t.Fatal(err)
	}

	if u.State != u2.State || u.Username != u2.Username {
		t.Fatalf("users are different: %q vs %q", u, u2)
	}
}

func TestPutParent(t *testing.T) {
	p := &Post{Content: "hello!"}
	if err := Put(ctx, p, WithParent(clark.Key)); err != nil {
		t.Fatal(err)
	}

	if *p.Key.Parent() != *clark.Key {
		t.Fatal("parent wasn't set")
	}
}

func TestDelete(t *testing.T) {
	u := &User{}
	Put(ctx, u)

	if err := Delete(ctx, u); err != nil {
		t.Fatal(err)
	}

	if u.Meta.State != EntityStateDeleted {
		t.Fatal("state was not updated")
	}

	if u.DeletedAt.IsZero() {
		t.Fatal("DeletedAt was not updated")
	}

	u2 := &User{}
	if err := Get(ctx, u2, WithKey(u.Key)); err != nil {
		t.Fatal(err)
	}

	if u.State != u2.State || u2.State != EntityStateDeleted {
		t.Fatalf("user wasn't deleted: %q", u2)
	}
}

func TestDeleteNoSave(t *testing.T) {
	u := &User{}
	err := Delete(ctx, u)
	if err != ErrMissingKey {
		t.Fatalf("expected ErrMissingKey, but got %v", err)
	}
}

func TestQuery(t *testing.T) {
	var us []User

	if _, err := Query("User").GetAll(ctx, &us); err != nil {
		t.Fatal(err)
	}

	for _, u := range us {
		if u.State != EntityStateActive {
			t.Errorf("entity is not active: %q", u)
		}
	}
}

func TestMain(m *testing.M) {
	flag.Parse()

	aectx, done, err := aetest.NewContext()
	if err != nil {
		log.Fatal(err)
	}
	defer done()

	ctx = aectx
	clark = &User{Username: "Clark"}
	if err := Put(ctx, clark, WithStringID(ctx, knownID)); err != nil {
		log.Fatal(err)
	}

	os.Exit(m.Run())
}
