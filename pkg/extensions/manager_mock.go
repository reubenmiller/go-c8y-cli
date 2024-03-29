// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package extensions

import (
	"github.com/reubenmiller/go-c8y-cli/v2/internal/ghrepo"
	"io"
	"sync"
)

// Ensure, that ExtensionManagerMock does implement ExtensionManager.
// If this is not the case, regenerate this file with moq.
var _ ExtensionManager = &ExtensionManagerMock{}

// ExtensionManagerMock is a mock implementation of ExtensionManager.
//
//	func TestSomethingThatUsesExtensionManager(t *testing.T) {
//
//		// make and configure a mocked ExtensionManager
//		mockedExtensionManager := &ExtensionManagerMock{
//			CreateFunc: func(name string, tmplType ExtTemplateType) error {
//				panic("mock out the Create method")
//			},
//			DispatchFunc: func(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (bool, error) {
//				panic("mock out the Dispatch method")
//			},
//			EnableDryRunModeFunc: func()  {
//				panic("mock out the EnableDryRunMode method")
//			},
//			ExecuteFunc: func(exe string, args []string, isBinary bool, stdin io.Reader, stdout io.Writer, stderr io.Writer) (bool, error) {
//				panic("mock out the Execute method")
//			},
//			InstallFunc: func(interfaceMoqParam ghrepo.Interface, s1 string, s2 string) error {
//				panic("mock out the Install method")
//			},
//			InstallLocalFunc: func(dir string, name string) error {
//				panic("mock out the InstallLocal method")
//			},
//			ListFunc: func() []Extension {
//				panic("mock out the List method")
//			},
//			RemoveFunc: func(name string) error {
//				panic("mock out the Remove method")
//			},
//			UpgradeFunc: func(name string, force bool) error {
//				panic("mock out the Upgrade method")
//			},
//		}
//
//		// use mockedExtensionManager in code that requires ExtensionManager
//		// and then make assertions.
//
//	}
type ExtensionManagerMock struct {
	// CreateFunc mocks the Create method.
	CreateFunc func(name string, tmplType ExtTemplateType) error

	// DispatchFunc mocks the Dispatch method.
	DispatchFunc func(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (bool, error)

	// EnableDryRunModeFunc mocks the EnableDryRunMode method.
	EnableDryRunModeFunc func()

	// ExecuteFunc mocks the Execute method.
	ExecuteFunc func(exe string, args []string, isBinary bool, stdin io.Reader, stdout io.Writer, stderr io.Writer) (bool, error)

	// InstallFunc mocks the Install method.
	InstallFunc func(interfaceMoqParam ghrepo.Interface, s1 string, s2 string) error

	// InstallLocalFunc mocks the InstallLocal method.
	InstallLocalFunc func(dir string, name string) error

	// ListFunc mocks the List method.
	ListFunc func() []Extension

	// RemoveFunc mocks the Remove method.
	RemoveFunc func(name string) error

	// UpgradeFunc mocks the Upgrade method.
	UpgradeFunc func(name string, force bool) error

	// calls tracks calls to the methods.
	calls struct {
		// Create holds details about calls to the Create method.
		Create []struct {
			// Name is the name argument value.
			Name string
			// TmplType is the tmplType argument value.
			TmplType ExtTemplateType
		}
		// Dispatch holds details about calls to the Dispatch method.
		Dispatch []struct {
			// Args is the args argument value.
			Args []string
			// Stdin is the stdin argument value.
			Stdin io.Reader
			// Stdout is the stdout argument value.
			Stdout io.Writer
			// Stderr is the stderr argument value.
			Stderr io.Writer
		}
		// EnableDryRunMode holds details about calls to the EnableDryRunMode method.
		EnableDryRunMode []struct {
		}
		// Execute holds details about calls to the Execute method.
		Execute []struct {
			// Exe is the exe argument value.
			Exe string
			// Args is the args argument value.
			Args []string
			// IsBinary is the isBinary argument value.
			IsBinary bool
			// Stdin is the stdin argument value.
			Stdin io.Reader
			// Stdout is the stdout argument value.
			Stdout io.Writer
			// Stderr is the stderr argument value.
			Stderr io.Writer
		}
		// Install holds details about calls to the Install method.
		Install []struct {
			// InterfaceMoqParam is the interfaceMoqParam argument value.
			InterfaceMoqParam ghrepo.Interface
			// S1 is the s1 argument value.
			S1 string
			// S2 is the s2 argument value.
			S2 string
		}
		// InstallLocal holds details about calls to the InstallLocal method.
		InstallLocal []struct {
			// Dir is the dir argument value.
			Dir string
			// Name is the name argument value.
			Name string
		}
		// List holds details about calls to the List method.
		List []struct {
		}
		// Remove holds details about calls to the Remove method.
		Remove []struct {
			// Name is the name argument value.
			Name string
		}
		// Upgrade holds details about calls to the Upgrade method.
		Upgrade []struct {
			// Name is the name argument value.
			Name string
			// Force is the force argument value.
			Force bool
		}
	}
	lockCreate           sync.RWMutex
	lockDispatch         sync.RWMutex
	lockEnableDryRunMode sync.RWMutex
	lockExecute          sync.RWMutex
	lockInstall          sync.RWMutex
	lockInstallLocal     sync.RWMutex
	lockList             sync.RWMutex
	lockRemove           sync.RWMutex
	lockUpgrade          sync.RWMutex
}

// Create calls CreateFunc.
func (mock *ExtensionManagerMock) Create(name string, tmplType ExtTemplateType) error {
	if mock.CreateFunc == nil {
		panic("ExtensionManagerMock.CreateFunc: method is nil but ExtensionManager.Create was just called")
	}
	callInfo := struct {
		Name     string
		TmplType ExtTemplateType
	}{
		Name:     name,
		TmplType: tmplType,
	}
	mock.lockCreate.Lock()
	mock.calls.Create = append(mock.calls.Create, callInfo)
	mock.lockCreate.Unlock()
	return mock.CreateFunc(name, tmplType)
}

// CreateCalls gets all the calls that were made to Create.
// Check the length with:
//
//	len(mockedExtensionManager.CreateCalls())
func (mock *ExtensionManagerMock) CreateCalls() []struct {
	Name     string
	TmplType ExtTemplateType
} {
	var calls []struct {
		Name     string
		TmplType ExtTemplateType
	}
	mock.lockCreate.RLock()
	calls = mock.calls.Create
	mock.lockCreate.RUnlock()
	return calls
}

// Dispatch calls DispatchFunc.
func (mock *ExtensionManagerMock) Dispatch(args []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) (bool, error) {
	if mock.DispatchFunc == nil {
		panic("ExtensionManagerMock.DispatchFunc: method is nil but ExtensionManager.Dispatch was just called")
	}
	callInfo := struct {
		Args   []string
		Stdin  io.Reader
		Stdout io.Writer
		Stderr io.Writer
	}{
		Args:   args,
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	}
	mock.lockDispatch.Lock()
	mock.calls.Dispatch = append(mock.calls.Dispatch, callInfo)
	mock.lockDispatch.Unlock()
	return mock.DispatchFunc(args, stdin, stdout, stderr)
}

// DispatchCalls gets all the calls that were made to Dispatch.
// Check the length with:
//
//	len(mockedExtensionManager.DispatchCalls())
func (mock *ExtensionManagerMock) DispatchCalls() []struct {
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
} {
	var calls []struct {
		Args   []string
		Stdin  io.Reader
		Stdout io.Writer
		Stderr io.Writer
	}
	mock.lockDispatch.RLock()
	calls = mock.calls.Dispatch
	mock.lockDispatch.RUnlock()
	return calls
}

// EnableDryRunMode calls EnableDryRunModeFunc.
func (mock *ExtensionManagerMock) EnableDryRunMode() {
	if mock.EnableDryRunModeFunc == nil {
		panic("ExtensionManagerMock.EnableDryRunModeFunc: method is nil but ExtensionManager.EnableDryRunMode was just called")
	}
	callInfo := struct {
	}{}
	mock.lockEnableDryRunMode.Lock()
	mock.calls.EnableDryRunMode = append(mock.calls.EnableDryRunMode, callInfo)
	mock.lockEnableDryRunMode.Unlock()
	mock.EnableDryRunModeFunc()
}

// EnableDryRunModeCalls gets all the calls that were made to EnableDryRunMode.
// Check the length with:
//
//	len(mockedExtensionManager.EnableDryRunModeCalls())
func (mock *ExtensionManagerMock) EnableDryRunModeCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockEnableDryRunMode.RLock()
	calls = mock.calls.EnableDryRunMode
	mock.lockEnableDryRunMode.RUnlock()
	return calls
}

// Execute calls ExecuteFunc.
func (mock *ExtensionManagerMock) Execute(exe string, args []string, isBinary bool, stdin io.Reader, stdout io.Writer, stderr io.Writer) (bool, error) {
	if mock.ExecuteFunc == nil {
		panic("ExtensionManagerMock.ExecuteFunc: method is nil but ExtensionManager.Execute was just called")
	}
	callInfo := struct {
		Exe      string
		Args     []string
		IsBinary bool
		Stdin    io.Reader
		Stdout   io.Writer
		Stderr   io.Writer
	}{
		Exe:      exe,
		Args:     args,
		IsBinary: isBinary,
		Stdin:    stdin,
		Stdout:   stdout,
		Stderr:   stderr,
	}
	mock.lockExecute.Lock()
	mock.calls.Execute = append(mock.calls.Execute, callInfo)
	mock.lockExecute.Unlock()
	return mock.ExecuteFunc(exe, args, isBinary, stdin, stdout, stderr)
}

// ExecuteCalls gets all the calls that were made to Execute.
// Check the length with:
//
//	len(mockedExtensionManager.ExecuteCalls())
func (mock *ExtensionManagerMock) ExecuteCalls() []struct {
	Exe      string
	Args     []string
	IsBinary bool
	Stdin    io.Reader
	Stdout   io.Writer
	Stderr   io.Writer
} {
	var calls []struct {
		Exe      string
		Args     []string
		IsBinary bool
		Stdin    io.Reader
		Stdout   io.Writer
		Stderr   io.Writer
	}
	mock.lockExecute.RLock()
	calls = mock.calls.Execute
	mock.lockExecute.RUnlock()
	return calls
}

// Install calls InstallFunc.
func (mock *ExtensionManagerMock) Install(interfaceMoqParam ghrepo.Interface, s1 string, s2 string) error {
	if mock.InstallFunc == nil {
		panic("ExtensionManagerMock.InstallFunc: method is nil but ExtensionManager.Install was just called")
	}
	callInfo := struct {
		InterfaceMoqParam ghrepo.Interface
		S1                string
		S2                string
	}{
		InterfaceMoqParam: interfaceMoqParam,
		S1:                s1,
		S2:                s2,
	}
	mock.lockInstall.Lock()
	mock.calls.Install = append(mock.calls.Install, callInfo)
	mock.lockInstall.Unlock()
	return mock.InstallFunc(interfaceMoqParam, s1, s2)
}

// InstallCalls gets all the calls that were made to Install.
// Check the length with:
//
//	len(mockedExtensionManager.InstallCalls())
func (mock *ExtensionManagerMock) InstallCalls() []struct {
	InterfaceMoqParam ghrepo.Interface
	S1                string
	S2                string
} {
	var calls []struct {
		InterfaceMoqParam ghrepo.Interface
		S1                string
		S2                string
	}
	mock.lockInstall.RLock()
	calls = mock.calls.Install
	mock.lockInstall.RUnlock()
	return calls
}

// InstallLocal calls InstallLocalFunc.
func (mock *ExtensionManagerMock) InstallLocal(dir string, name string) error {
	if mock.InstallLocalFunc == nil {
		panic("ExtensionManagerMock.InstallLocalFunc: method is nil but ExtensionManager.InstallLocal was just called")
	}
	callInfo := struct {
		Dir  string
		Name string
	}{
		Dir:  dir,
		Name: name,
	}
	mock.lockInstallLocal.Lock()
	mock.calls.InstallLocal = append(mock.calls.InstallLocal, callInfo)
	mock.lockInstallLocal.Unlock()
	return mock.InstallLocalFunc(dir, name)
}

// InstallLocalCalls gets all the calls that were made to InstallLocal.
// Check the length with:
//
//	len(mockedExtensionManager.InstallLocalCalls())
func (mock *ExtensionManagerMock) InstallLocalCalls() []struct {
	Dir  string
	Name string
} {
	var calls []struct {
		Dir  string
		Name string
	}
	mock.lockInstallLocal.RLock()
	calls = mock.calls.InstallLocal
	mock.lockInstallLocal.RUnlock()
	return calls
}

// List calls ListFunc.
func (mock *ExtensionManagerMock) List() []Extension {
	if mock.ListFunc == nil {
		panic("ExtensionManagerMock.ListFunc: method is nil but ExtensionManager.List was just called")
	}
	callInfo := struct {
	}{}
	mock.lockList.Lock()
	mock.calls.List = append(mock.calls.List, callInfo)
	mock.lockList.Unlock()
	return mock.ListFunc()
}

// ListCalls gets all the calls that were made to List.
// Check the length with:
//
//	len(mockedExtensionManager.ListCalls())
func (mock *ExtensionManagerMock) ListCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockList.RLock()
	calls = mock.calls.List
	mock.lockList.RUnlock()
	return calls
}

// Remove calls RemoveFunc.
func (mock *ExtensionManagerMock) Remove(name string) error {
	if mock.RemoveFunc == nil {
		panic("ExtensionManagerMock.RemoveFunc: method is nil but ExtensionManager.Remove was just called")
	}
	callInfo := struct {
		Name string
	}{
		Name: name,
	}
	mock.lockRemove.Lock()
	mock.calls.Remove = append(mock.calls.Remove, callInfo)
	mock.lockRemove.Unlock()
	return mock.RemoveFunc(name)
}

// RemoveCalls gets all the calls that were made to Remove.
// Check the length with:
//
//	len(mockedExtensionManager.RemoveCalls())
func (mock *ExtensionManagerMock) RemoveCalls() []struct {
	Name string
} {
	var calls []struct {
		Name string
	}
	mock.lockRemove.RLock()
	calls = mock.calls.Remove
	mock.lockRemove.RUnlock()
	return calls
}

// Upgrade calls UpgradeFunc.
func (mock *ExtensionManagerMock) Upgrade(name string, force bool) error {
	if mock.UpgradeFunc == nil {
		panic("ExtensionManagerMock.UpgradeFunc: method is nil but ExtensionManager.Upgrade was just called")
	}
	callInfo := struct {
		Name  string
		Force bool
	}{
		Name:  name,
		Force: force,
	}
	mock.lockUpgrade.Lock()
	mock.calls.Upgrade = append(mock.calls.Upgrade, callInfo)
	mock.lockUpgrade.Unlock()
	return mock.UpgradeFunc(name, force)
}

// UpgradeCalls gets all the calls that were made to Upgrade.
// Check the length with:
//
//	len(mockedExtensionManager.UpgradeCalls())
func (mock *ExtensionManagerMock) UpgradeCalls() []struct {
	Name  string
	Force bool
} {
	var calls []struct {
		Name  string
		Force bool
	}
	mock.lockUpgrade.RLock()
	calls = mock.calls.Upgrade
	mock.lockUpgrade.RUnlock()
	return calls
}
