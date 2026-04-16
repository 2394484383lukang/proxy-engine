package proxy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectProxyDial(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("hello"))
	}))
	defer ts.Close()

	d := NewDirectProxy()
	conn, err := d.Dial(context.Background(), ts.Listener.Addr().String())
	assert.NoError(t, err)
	assert.NotNil(t, conn)
	conn.Close()
}

func TestDirectProxyType(t *testing.T) {
	d := NewDirectProxy()
	assert.Equal(t, ProxyTypeDirect, d.Type())
}

func TestRejectProxyDial(t *testing.T) {
	r := &RejectProxy{}
	conn, err := r.Dial(context.Background(), "example.com:80")
	assert.Error(t, err)
	assert.Nil(t, conn)
	assert.Equal(t, ErrRejected, err)
}

func TestRejectProxyType(t *testing.T) {
	r := &RejectProxy{}
	assert.Equal(t, ProxyTypeReject, r.Type())
}
