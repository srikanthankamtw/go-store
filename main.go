package main

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"net/http"
	"sync"
)

type Storer[K comparable, V any] interface {
	Create(K, V) error
	Read(K) (V, error)
	Update(K, V) error
	Delete(K) (V, error)
}

type KVStore[K comparable, V any] struct {
	mu   sync.RWMutex
	data map[K]V
}

func NewKVStore[K comparable, V any]() *KVStore[K, V] {
	return &KVStore[K, V]{
		data: make(map[K]V),
	}
}

// Has checks if the given key is present in the store
// NOTE: this is not concurrent safe, should use with a lock or mutex
func (k *KVStore[K, V]) Has(key K) (V, bool) {
	value, ok := k.data[key]
	return value, ok
}

func (k *KVStore[K, V]) Create(key K, value V) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	k.data[key] = value
	return nil
}

func (k *KVStore[K, V]) Read(key K) (V, error) {
	k.mu.RLock()
	defer k.mu.RUnlock()
	value, ok := k.data[key]
	if !ok {
		return value, fmt.Errorf("key (%v) does not exists", key)
	}
	return value, nil
}

func (k *KVStore[K, V]) Update(key K, value V) error {
	k.mu.Lock()
	defer k.mu.Unlock()
	_, keyExists := k.Has(key)
	if keyExists {
		k.data[key] = value
	}
	return nil
}

func (k *KVStore[K, V]) Delete(key K) (V, error) {
	k.mu.Lock()
	defer k.mu.Unlock()
	v, keyExists := k.Has(key)
	if keyExists {
		delete(k.data, key)
	}
	return v, nil
}

type Server struct {
	Store Storer[string, string]
	Port  string
}

func NewServer(port string) *Server {
	return &Server{
		Store: NewKVStore[string, string](),
		Port:  port,
	}
}

func (s *Server) handleCreate(c echo.Context) error {
	key := c.Param("key")
	value := c.Param("value")
	err := s.Store.Create(key, value)
	if err != nil {
		return err
	}
	err = c.JSON(http.StatusOK, map[string]string{"message": "created"})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) handleRead(c echo.Context) error {
	key := c.Param("key")
	value, err := s.Store.Read(key)
	if err != nil {
		return err
	}
	err = c.JSON(http.StatusOK, map[string]string{"value": value})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) handleUpdate(c echo.Context) error {
	key := c.Param("key")
	value := c.Param("value")
	err := s.Store.Update(key, value)
	if err != nil {
		return err
	}
	err = c.JSON(http.StatusOK, map[string]string{"message": "updated", "key": key, "value": value})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) handleDelete(c echo.Context) error {
	key := c.Param("key")
	value, err := s.Store.Delete(key)
	if err != nil {
		return err
	}
	err = c.JSON(http.StatusOK, map[string]string{"message": "deleted", "value": value})
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) Start() {
	fmt.Printf("HTTP Server is running on port %s", s.Port)
	e := echo.New()

	e.GET("/create/:key/:value", s.handleCreate)
	e.GET("/read/:key", s.handleRead)
	e.GET("/update/:key/:value", s.handleUpdate)
	e.GET("/delete/:key", s.handleDelete)

	_ = e.Start(s.Port)
}

func main() {
	server := NewServer(":3000")
	server.Start()
}
