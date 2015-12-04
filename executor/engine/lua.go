package engine

import (
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/Shopify/go-lua"
	"golang.org/x/net/context"
)

type LuaOption func(*LuaProgram)

func SetMetricReporter(met MetricReporter) LuaOption {
	return func(prgm *LuaProgram) {
		prgm.metrics = met
	}
}

func SetLogger(out io.Writer) LuaOption {
	return func(prgm *LuaProgram) {
		prgm.out = out
	}
}

var _ Program = &LuaProgram{}

type LuaProgram struct {
	vm *lua.State

	// program state
	configured bool
	steps      []string

	metrics MetricReporter
	info    func(*lua.State) int
	fatal   func(*lua.State) int

	out io.Writer
}

func Lua(source io.Reader, opts ...LuaOption) (*LuaProgram, error) {
	l := lua.NewState()

	prgm := &LuaProgram{
		vm:      l,
		out:     ioutil.Discard,
		metrics: nullMetric{},
		info: func(l *lua.State) int {
			panic(fmt.Errorf("'info' is not defined outside of steps"))
			return 0
		},
		fatal: func(l *lua.State) int {
			panic(fmt.Errorf("'fatal' is not defined outside of steps"))
			return 0
		},
	}

	for _, opt := range opts {
		opt(prgm)
	}

	// setup the `step` hooksA
	configureLua(prgm, l)

	l.Register("info", func(l *lua.State) int {
		prgm.metrics.IncrLogInfo(l.ToValue(1))
		return prgm.info(l)
	})
	l.Register("fatal", func(l *lua.State) int {
		prgm.metrics.IncrLogFatal(l.ToValue(1))
		return prgm.fatal(l)
	})

	httpBind := newHTTPBinding(prgm.metrics)
	l.Register("get", httpBind.get)
	l.Register("post", httpBind.post)

	// load the source
	if err := l.Load(source, "", ""); err != nil {
		return prgm, fmt.Errorf("compiling program: %v", err)
	}
	// invoke the program to prepare the steps
	if err := l.ProtectedCall(0, 0, 0); err != nil {
		return prgm, fmt.Errorf("preparing program: %v", err)
	}

	prgm.configured = true

	return prgm, nil
}

func VerifyConfig(json map[string]interface{}) error {
	for key, val := range json {
		switch valType := val.(type) {
		case string:
		case int:
		case float64:
		case bool:
		case nil:
			break
		default:
			return fmt.Errorf("invalid type for lua config file: %v, key was: %v", valType, key)
		}
	}
	return nil
}

func (prgm *LuaProgram) AddConfig(json map[string]interface{}) error {
	vm := prgm.vm
	for key, val := range json {
		switch valType := val.(type) {
		case string:
			vm.PushString(fmt.Sprintf("%v", val))
			vm.SetGlobal(key)
		case int:
			vm.PushString(fmt.Sprintf("%v", val))
			vm.SetGlobal(key)
		case float64:
			vm.PushString(fmt.Sprintf("%v", val))
			vm.SetGlobal(key)
		case bool:
			vm.PushString(fmt.Sprintf("%v", val))
			vm.SetGlobal(key)
		case nil:
			vm.PushNil()
			vm.SetGlobal(key)
		default:
			return fmt.Errorf("invalid type for lua config file: %v, key was: %v", valType, key)
		}
	}
	return nil
}

func (prgm *LuaProgram) Execute(ctx context.Context) error {

	runNext := true
	currentStep := "<not a step>"

	prgm.info = func(l *lua.State) int {
		msg := l.ToValue(1)
		fmt.Fprintf(prgm.out, `{"lvl":"info","step":%q,"msg":"%v"}`+"\n", currentStep, msg)
		return 0
	}
	prgm.fatal = func(l *lua.State) int {
		msg := l.ToValue(1)
		fmt.Fprintf(prgm.out, `{"lvl":"fatal","step":%q,"msg":"%v"}`+"\n", currentStep, msg)
		return 0
	}

	reporter := func(stepName string) bool {
		currentStep = stepName
		select {
		default:
			return runNext
		case <-ctx.Done():
			return false
		}
	}

	return prgm.runSteps(reporter)
}

func (prgm *LuaProgram) runSteps(reporter func(step string) bool) error {
	l := prgm.vm

	// bring the step table on the stack
	l.Global("step")
	// prepare a table to hold the results of calling
	// each step
	l.NewTable()
	defer l.Pop(2) // cleanup the step + table

	prgm.metrics.IncrScriptExecution()
	for _, stepName := range prgm.steps {
		if !reporter(stepName) {
			// stop running
			break
		}

		i := l.Top()
		// pull the func at `stepName` out of the table
		l.Field(-i, stepName)
		// copy the table that holds results of the last step
		// as an argument for the func about to be invoked
		l.PushValue(-i)
		// remove the old copy of the result
		l.Remove(-(i + 1))
		// now that we have the:
		//   - argument
		//   - function
		//   - step-table
		// we can invoke the function with the argument

		start := time.Now()
		err := l.ProtectedCall(1, 1, 0) // 1 argument, with 1 return value
		if err != nil {
			prgm.metrics.IncrStepError(stepName)
			return &StepError{Step: stepName, Err: err}
		}
		if l.Top() != 2 {
			lua.Errorf(l, "step %q needs to return exactly 1 argument", stepName)
		}
		prgm.metrics.IncrStepExecution(stepName, time.Since(start))
	}
	return nil
}

func (prgm *LuaProgram) registerStep(l *lua.State) int {
	if prgm.configured {
		lua.Errorf(l, "step is immutable")
		return 0
	}
	stepName := lua.CheckString(l, 2)
	l.RawSet(1)
	prgm.steps = append(prgm.steps, stepName)
	return 0
}

func configureLua(prgm *LuaProgram, l *lua.State) {
	lua.NewMetaTable(l, "stepMetaTable")
	lua.SetFunctions(l, []lua.RegistryFunction{{
		"__newindex", prgm.registerStep,
	}}, 0)

	// create the `step` table, make it global, give it the
	// meta table that intercepts `__newindex` and then pop
	// them off the stack
	l.NewTable()
	l.PushValue(-1)
	l.SetGlobal("step")
	lua.SetMetaTableNamed(l, "stepMetaTable")
	l.Pop(2)
}
