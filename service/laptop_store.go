// Copyright 2024 Emory.Du <orangeduxiaocheng@gmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/emorydu/grpc-examples/pb"
	"github.com/jinzhu/copier"
	"log"
	"sync"
)

// ErrAlreadyExists is returned when a record with the same ID already exists in the store
var ErrAlreadyExists = errors.New("record already exists")

// LaptopStore is an interface to store laptop
type LaptopStore interface {
	// Save saves the laptop to the store
	Save(laptop *pb.Laptop) error
	// Find finds a laptop by ID
	Find(id string) (*pb.Laptop, error)
	// Search searches for laptops with filter, returns one by one via the found function
	Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error
}

// InMemoryLaptopStore stores laptop in memory
type InMemoryLaptopStore struct {
	sync.RWMutex

	data map[string]*pb.Laptop
}

// NewInMemoryLaptopStore returns a new InMemoryLaptopStore
func NewInMemoryLaptopStore() *InMemoryLaptopStore {
	return &InMemoryLaptopStore{
		data: make(map[string]*pb.Laptop),
	}
}

// Save saves the laptop to the store
func (m *InMemoryLaptopStore) Save(laptop *pb.Laptop) error {
	m.Lock()
	defer m.Unlock()

	if m.data[laptop.Id] != nil {
		return ErrAlreadyExists
	}

	other, err := deepCopy(laptop)
	if err != nil {
		return err
	}

	m.data[other.Id] = other

	return nil
}

// Find finds a laptop by ID
func (m *InMemoryLaptopStore) Find(id string) (*pb.Laptop, error) {
	m.RLock()
	defer m.RUnlock()

	laptop := m.data[id]

	if laptop == nil {
		return nil, nil
	}

	return deepCopy(laptop)
}

// Search searches for laptops with filter, returns one by one via the found function
func (m *InMemoryLaptopStore) Search(ctx context.Context, filter *pb.Filter, found func(laptop *pb.Laptop) error) error {
	m.RLock()
	defer m.RUnlock()

	for _, laptop := range m.data {
		// heavy processing
		//time.Sleep(time.Second)
		//log.Print("checking laptop id: ", laptop.GetId())

		if err := ctx.Err(); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				log.Println("context is cancelled")
				return errors.New("context is cancelled")
			}
		}

		if isQualified(filter, laptop) {
			other, err := deepCopy(laptop)
			if err != nil {
				return err
			}
			err = found(other)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func isQualified(filter *pb.Filter, laptop *pb.Laptop) bool {
	if laptop.GetPriceUsd() > filter.GetMaxPriceUsd() {
		return false
	}
	if laptop.GetCpu().GetNumberCores() < filter.GetMinCpuCores() {
		return false
	}
	if laptop.GetCpu().GetMinGhz() < filter.GetMinCpuGhz() {
		return false
	}
	if toBit(laptop.GetRam()) < toBit(filter.GetMinRam()) {
		return false
	}

	return true
}

func toBit(memory *pb.Memory) uint64 {
	value := memory.GetValue()

	switch memory.GetUnit() {
	case pb.Memory_BIT:
		return value
	case pb.Memory_BYTE:
		return value << 3 // 8 = 2^3
	case pb.Memory_KILOBYTE:
		return value << 13 // 1024 * 8 = 2^10 * 2^3
	case pb.Memory_MEGABYTE:
		return value << 23
	case pb.Memory_GIGABYTE:
		return value << 33
	case pb.Memory_TERABYTE:
		return value << 43
	default:
		return 0
	}
}

func deepCopy(laptop *pb.Laptop) (*pb.Laptop, error) {
	// deep copy
	other := &pb.Laptop{}
	err := copier.Copy(other, laptop)
	if err != nil {
		return nil, fmt.Errorf("cannot copy laptop data: %w", err)
	}

	return other, nil
}
