package glib

// #include <glib.h>
// #include <glib-object.h>
// #include "glib.go.h"
import "C"
import (
	"reflect"
	"unsafe"

	"github.com/diamondburned/go-glib/core/closure"
	"github.com/diamondburned/go-glib/core/intern"
)

/*
 * Events
 */

// SignalHandle is the ID of a signal handler.
type SignalHandle uint

// Connect is a wrapper around g_signal_connect_closure(). f must be a function
// with at least one parameter matching the type it is connected to.
//
// It is optional to list the rest of the required types from Gtk, as values
// that don't fit into the function parameter will simply be ignored; however,
// extraneous types will trigger a runtime panic. Arguments for f must be a
// matching Go equivalent type for the C callback, or an interface type which
// the value may be packed in. If the type is not suitable, a runtime panic will
// occur when the signal is emitted.
//
// Circular References
//
// To prevent circular references, prefer declaring Connect functions like so:
//
//    obj.Connect(func(obj *ObjType) { obj.Do() })
//
// Instead of directly referencing the object from outside like so:
//
//    obj.Connect(func() { obj.Do() })
//
// When using Connect, beware of referencing variables outside the closure that
// may cause a circular reference that prevents both Go from garbage collecting
// the callback and GTK from successfully unreferencing its values.
//
// Below is an example piece of code that is considered "leaky":
//
//    type ChatBox struct {
//        gtk.TextView
//        Loader *gdk.PixbufLoader
//
//        State State
//    }
//
//    func (box *ChatBox) Method() {
//        box.Loader.Connect("size-allocate", func(loader *gdk.PixbufLoader) {
//            // Here, we're dereferencing box to get the state, which might
//            // keep box alive along with the PixbufLoader, causing a circular
//            // reference.
//            loader.SetSize(box.State.Width, box.State.Height)
//        })
//    }
//
// There are many solutions to fix the above piece of code. For example,
// box.Loader could be discarded manually immediately after it's done by setting
// it to nil, or the signal handle could be disconnected manually, or box could
// be set to nil after its first call in the callback.
func (v *Object) Connect(detailedSignal string, f interface{}) SignalHandle {
	return v.connectClosure(false, detailedSignal, f)
}

// ConnectAfter is a wrapper around g_signal_connect_closure(). The difference
// between Connect and ConnectAfter is that the latter will be invoked after the
// default handler, not before. For more information, refer to Connect.
func (v *Object) ConnectAfter(detailedSignal string, f interface{}) SignalHandle {
	return v.connectClosure(true, detailedSignal, f)
}

// ClosureCheckReceiver, if true, will make GLib check for every single
// closure's first argument to ensure that it is correct, otherwise it will
// panic with a message warning about the possible circular references. The
// receiver in this case is most often the first argument of the callback.
//
// This constant can be changed by using go.mod's replace directive for
// debugging purposes.
const ClosureCheckReceiver = false

func (v *Object) connectClosure(after bool, detailedSignal string, f interface{}) SignalHandle {
	fs := closure.NewFuncStack(f, 2)

	if ClosureCheckReceiver {
		// This is a bit slow, but we could be careful.
		objValue, err := v.goValue()
		if err == nil {
			fsType := fs.Func.Type()
			if fsType.NumIn() < 1 {
				fs.Panicf("callback should have the object receiver to avoid circular references")
			}
			objType := reflect.TypeOf(objValue)
			if first := fsType.In(0); !objType.ConvertibleTo(first) {
				fs.Panicf("receiver not convertible to expected type %s, got %s", objType, first)
			}
		}

		// Allow the type check to fail if we can't get a value marshaler. This
		// rarely happens, but it might, and we want to at least allow working
		// around it.
	}

	cstr := C.CString(detailedSignal)
	defer C.free(unsafe.Pointer(cstr))

	gclosure := v.ClosureNew(fs)
	c := C.g_signal_connect_closure(C.gpointer(v.GObject), (*C.gchar)(cstr), gclosure, gbool(after))

	return SignalHandle(c)
}

// ClosureNew creates a new GClosure that's bound to the current object and adds
// its callback function to the internal registry. It's exported for visibility
// to other gotk3 packages and should not be used in a regular application.
func (v *Object) ClosureNew(f interface{}) *C.GClosure {
	fs, ok := f.(*closure.FuncStack)
	if !ok {
		fs = closure.NewFuncStack(f, 2)
	}

	gclosure := C.g_closure_new_simple(C.sizeof_GClosure, nil)
	v.box.Closures.Register(unsafe.Pointer(gclosure), fs)

	C.g_closure_set_meta_marshal(gclosure, C.gpointer(v.GObject), (*[0]byte)(C.goMarshal))
	C.g_closure_add_finalize_notifier(gclosure, C.gpointer(v.GObject), (*[0]byte)(C.removeClosure))

	return gclosure
}

//export removeClosure
func removeClosure(obj *C.GObject, gclosure *C.GClosure) {
	intern.RemoveClosure(unsafe.Pointer(obj), unsafe.Pointer(gclosure))
}
