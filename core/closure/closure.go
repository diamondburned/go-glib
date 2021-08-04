package closure

import (
	"sync"
	"unsafe"
)

// Registry describes the local closure registry of each object.
type Registry struct {
	reg sync.Map // unsafe.Pointer(*C.GClosure) -> *FuncStack
}

// NewRegistry creates an empty closure registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// Register registers the given GClosure callback.
func (r *Registry) Register(gclosure unsafe.Pointer, callback *FuncStack) {
	r.reg.Store(gclosure, callback)
}

// Load loads the given GClosure's callback. Nil is returned if it's not found.
func (r *Registry) Load(gclosure unsafe.Pointer) *FuncStack {
	fs, ok := r.reg.Load(gclosure)
	if !ok {
		return nil
	}
	return fs.(*FuncStack)
}

// Delete deletes the given GClosure callback.
func (r *Registry) Delete(gclosure unsafe.Pointer) {
	r.reg.Delete(gclosure)
}

/*
var (
	closures = sync.Map{} // unsafe.Pointer(*GClosure) -> reflect.Value

	// use a bi-directional map to allow lookup of the closure value from both
	// the SourceHandle and the closure ID in constant time.
	signalMu       sync.Mutex
	signalClosures = map[uint]unsafe.Pointer{} // uint(SourceHandle) -> unsafe.Pointer (closure key, callbackID)
	closureSignals = map[unsafe.Pointer]uint{} // unsafe.Pointer(*GClosure) -> uint(SourceHandle)
)

// RegisterSignal registers the given signal handle to be associated with the
// closure pointer. This association allows the closure to be removed as well
// when the signal removal is requested from the user using DisconnectSignal.
func RegisterSignal(handle uint, closure unsafe.Pointer) {
	// Safety check omitted until the race condition in glib/connect.go is
	// fixed. Check that file for more info.

	signalMu.Lock()
	defer signalMu.Unlock()

	signalClosures[handle] = closure
	closureSignals[closure] = handle
}

// DisconnectSignal removes both the signal and the closure associated with it
// from the internal registry. Since this function will also remove the closure
// itself from the internal registry, Gtk's disconnect functions should be
// called first.
func DisconnectSignal(handle uint) {
	signalMu.Lock()
	defer signalMu.Unlock()

	closure, ok := signalClosures[handle]
	if ok {
		closures.Delete(closure)
		delete(closureSignals, closure)
		delete(signalClosures, handle)
	}
}

// Assign assigns the given FuncStack to the given closure.
func Assign(closure unsafe.Pointer, fs FuncStack) {
	closures.Store(closure, fs)
}

// Get gets the reflect-value callback from the closure pointer.
func Get(closure unsafe.Pointer) FuncStack {
	v, ok := closures.Load(closure)
	if ok {
		return v.(FuncStack)
	}
	return zeroFuncStack
}

// Delete deletes the closure pointer from the registry while also checking for
// any existing signal handler associated with the given callback ID. If a
// signal handler is found, then its behavior is similar to DisconnectSignal.
func Delete(closure unsafe.Pointer) {
	funcStack := getAndDeleteClosure(closure)
	if !funcStack.IsValid() {
		return
	}

	signalMu.Lock()
	defer signalMu.Unlock()

	handle, ok := closureSignals[closure]
	if ok {
		delete(closureSignals, closure)
		delete(signalClosures, handle)
	}
}
*/
