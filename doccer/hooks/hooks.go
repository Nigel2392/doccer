package hooks

import (
	"fmt"
	"reflect"
	"slices"
)

var DefaultRegistry = make(HookRegistry)

type _Hook struct {
	NumArgs  int
	Variadic bool
	VFunc    reflect.Value
	TFunc    reflect.Type
	Order    int
	Func     interface{}
}

// Create a new hook
func NewHook(order int, f interface{}) *_Hook {
	if h, ok := f.(*_Hook); ok {
		return h
	}

	if h, ok := f.(_Hook); ok {
		return &h
	}

	var (
		vFunc = reflect.ValueOf(f)
		tFunc = vFunc.Type()
	)

	if tFunc.Kind() != reflect.Func {
		panic(fmt.Sprintf("expected function, got %T", f))
	}

	return &_Hook{
		NumArgs:  tFunc.NumIn(),
		Variadic: tFunc.IsVariadic(),
		VFunc:    vFunc,
		TFunc:    tFunc,
		Order:    order,
		Func:     f,
	}
}

// Execute the hook function
func (h *_Hook) Call(args ...interface{}) (value interface{}, err error) {
	var (
		numArgs = len(args)
		values  = make([]reflect.Value, numArgs)
	)

	if h.Variadic {
		if numArgs < h.NumArgs-1 {
			return value, fmt.Errorf("expected at least %d arguments, got %d", h.NumArgs-1, numArgs)
		}
	} else if numArgs != h.NumArgs {
		return value, fmt.Errorf("expected %d arguments, got %d", h.NumArgs, numArgs)
	}

	for i, arg := range args {
		values[i] = reflect.ValueOf(arg)
	}

	var (
		results = h.VFunc.Call(values)
		r       = make([]interface{}, len(results))
	)
	for i, result := range results {
		r[i] = result.Interface()
	}
	if len(r) == 1 {
		return r[0], nil
	}
	return r, nil
}

type HookRegistry map[string][]*_Hook

func (h HookRegistry) Register(identifier string, order int, hooks ...interface{}) {
	var hooksList, ok = h[identifier]
	if !ok {
		hooksList = make([]*_Hook, 0)
	}
	for _, hook := range hooks {
		hooksList = append(hooksList, NewHook(order, hook))
	}
	h[identifier] = hooksList
}

// Register a new hook
func Register(identifier string, order int, hooks ...interface{}) {
	DefaultRegistry.Register(identifier, order, hooks...)
}

// Get the hooks
func Get[T any](identifier string) (h []T) {
	return get[T](DefaultRegistry, identifier)
}

// Get the hooks from a registry
func GetFrom[T any](registry HookRegistry, identifier string) (h []T) {
	return get[T](registry, identifier)
}

// Get the hooks, casting the interfaces back to functions
func get[T any](registry HookRegistry, identifier string) (h []T) {
	var hooks, ok = registry[identifier]
	if !ok {
		return make([]T, 0)
	}

	// Sort the hooks by order.
	slices.SortFunc(hooks, func(i, j *_Hook) int {
		// Sort by order, the higher the order the later it is called
		if i.Order < j.Order {
			return -1
		} else if i.Order > j.Order {
			return 1
		}
		return 0
	})

	// Cast the hooks back to function types
	var (
		hookList = make([]T, len(hooks))
		typeOfT  = reflect.TypeOf(*new(T))
	)
	for i, hook := range hooks {
		if hook.TFunc.ConvertibleTo(typeOfT) {
			hookList[i] = hook.VFunc.Convert(typeOfT).Interface().(T)
		} else {
			panic(fmt.Sprintf("hook %d is not of type %T but %T, cannot convert.", i, hook.Func, hook.TFunc))
		}
	}

	return hookList
}
