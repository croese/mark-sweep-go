package main

import (
	"fmt"
)

type objectType int

const (
	intType            = objectType(0)
	pairType           = objectType(1)
	maxStackSize       = 256
	initialGCThreshold = 10
)

type object struct {
	marked   bool
	typeFlag objectType
	value    int
	head     *object
	tail     *object
	next     *object
}

func (o *object) String() string {
	switch o.typeFlag {
	case intType:
		return fmt.Sprintf("%d", o.value)
	case pairType:
		return fmt.Sprintf("(%s, %s)", o.head, o.tail)
	default:
		panic("Unknown type")
	}
}

type vm struct {
	stack       []*object
	stackSize   int
	firstObject *object
	numObjects  int
	maxObjects  int
}

func main() {
	vm := newVM(maxStackSize, initialGCThreshold)
	vm.pushInt(42)
	i := vm.newObject(intType)
	i.value = 5
	vm.pushInt(1)
	vm.pushInt(2)
	vm.dump()
	vm.pushPair()
	vm.dump()
	vm.pushPair()
	i = vm.newObject(intType)
	i.value = 123
	vm.dump()
	vm.gc()
	vm.dump()
}

func (v *vm) dump() {
	fmt.Println("\n---Start Stack Dump---")
	for i := v.stackSize - 1; i >= 0; i-- {
		fmt.Printf("%d: %s\n", i, v.stack[i])
	}
	fmt.Println("---End Stack Dump---")

	fmt.Println("\n---Start Memory Dump---")
	fmt.Printf("Num. of objects: %d\n", v.numObjects)
	fmt.Printf("Max objects: %d\n", v.maxObjects)
	for o := v.firstObject; o != nil; o = o.next {
		fmt.Printf("%p: %s\n", o, o)
	}
	fmt.Println("---End Memory Dump---")
}

func newVM(maxStackSize int, gcThreshold int) *vm {
	return &vm{
		stack:       make([]*object, maxStackSize),
		stackSize:   0,
		firstObject: nil,
		numObjects:  0,
		maxObjects:  gcThreshold,
	}
}

func (v *vm) newObject(t objectType) *object {
	if v.numObjects == v.maxObjects {
		v.gc()
	}

	o := &object{
		typeFlag: t,
		marked:   false,
	}

	o.next = v.firstObject
	v.firstObject = o

	v.numObjects++

	return o
}

func (v *vm) push(value *object) {
	if v.stackSize >= len(v.stack) {
		panic("Stack overflow")
	}

	v.stack[v.stackSize] = value
	v.stackSize++
}

func (v *vm) pop() *object {
	if v.stackSize == 0 {
		panic("Stack underflow")
	}

	idx := v.stackSize - 1
	top := v.stack[idx]
	v.stack[idx] = nil
	v.stackSize--

	return top
}

func (v *vm) pushInt(value int) {
	o := v.newObject(intType)
	o.value = value
	v.push(o)
}

func (v *vm) popInt() int {
	o := v.pop()
	if o.typeFlag == intType {
		return o.value
	}

	panic("Top stack value was not an int")
}

func (v *vm) pushPair() {
	o := v.newObject(pairType)
	o.tail = v.pop()
	o.head = v.pop()

	v.push(o)
}

func (v *vm) popPair() *object {
	o := v.pop()
	if o.typeFlag == pairType {
		return o
	}

	panic("Top stack value was not a pair")
}

func (v *vm) add() {
	right := v.popInt()
	left := v.popInt()

	v.pushInt(left + right)
}

func mark(o *object) {
	if o.marked {
		return
	}

	o.marked = true

	if o.typeFlag == pairType {
		mark(o.head)
		mark(o.tail)
	}
}

func (v *vm) markAll() {
	fmt.Println("\n---Start Mark Phase---")
	for i := 0; i < v.stackSize; i++ {
		mark(v.stack[i])
	}
	fmt.Println("---End Mark Phase---")
}

func (v *vm) sweep() {
	fmt.Println("\n---Start Sweep Phase---")
	o := v.firstObject
	var prev *object
	for o != nil {
		if !o.marked {
			unreached := o
			if prev == nil {
				// first object
				v.firstObject = unreached.next
			} else {
				// anything after first
				prev.next = unreached.next
			}
			o = unreached.next
			v.numObjects--
			fmt.Printf("Object freed: %s\n", unreached)
		} else {
			o.marked = false
			prev = o
			o = o.next
		}
	}
	fmt.Println("---End Sweep Phase---")
}

func (v *vm) gc() {
	fmt.Println("\n---Start GC---")
	v.markAll()
	v.sweep()
	fmt.Println("\n---End GC---")

	v.maxObjects = v.numObjects * 2
}
