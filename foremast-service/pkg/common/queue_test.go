package common

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestQueue(t *testing.T) {
	// get instance should not return nil
	q := GetQueueInstance()
	assert.NotNil(t, q)
	// initial queue should empty
	assert.Equal(t,0, q.Len())
	obj, _ := q.Front()
	assert.Nil(t, obj)
	obj, _ = q.Pop()
	assert.Nil(t, obj)
	// add objects
	q.Push("obj1")
	q.Push("obj2")
	// top should be obj1
	front, _ := q.Front()
	assert.Equal(t,"obj1", front)
	// size still 2
	assert.Equal(t, 2, q.Len())
	// get and remove top obj
	obj, _ = q.Pop()
	assert.Equal(t, "obj1", obj)
	assert.Equal(t, 1, q.Len())
	// queue initial slice size is 16
	assert.Equal(t, 16, q.lastSliceSize)
	for i :=0; i < 16; i++ {
		q.Push(i)
	}
	assert.Equal(t, 128, q.lastSliceSize)
}
