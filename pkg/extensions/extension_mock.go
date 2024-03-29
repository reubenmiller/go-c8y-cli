// Code generated by moq; DO NOT EDIT.
// github.com/matryer/moq

package extensions

import (
	"sync"
)

// Ensure, that ExtensionMock does implement Extension.
// If this is not the case, regenerate this file with moq.
var _ Extension = &ExtensionMock{}

// ExtensionMock is a mock implementation of Extension.
//
//	func TestSomethingThatUsesExtension(t *testing.T) {
//
//		// make and configure a mocked Extension
//		mockedExtension := &ExtensionMock{
//			AliasesFunc: func() ([]Alias, error) {
//				panic("mock out the Aliases method")
//			},
//			CommandsFunc: func() ([]Command, error) {
//				panic("mock out the Commands method")
//			},
//			CurrentVersionFunc: func() string {
//				panic("mock out the CurrentVersion method")
//			},
//			IsBinaryFunc: func() bool {
//				panic("mock out the IsBinary method")
//			},
//			IsLocalFunc: func() bool {
//				panic("mock out the IsLocal method")
//			},
//			IsPinnedFunc: func() bool {
//				panic("mock out the IsPinned method")
//			},
//			NameFunc: func() string {
//				panic("mock out the Name method")
//			},
//			PathFunc: func() string {
//				panic("mock out the Path method")
//			},
//			TemplatePathFunc: func() string {
//				panic("mock out the TemplatePath method")
//			},
//			URLFunc: func() string {
//				panic("mock out the URL method")
//			},
//			UpdateAvailableFunc: func() bool {
//				panic("mock out the UpdateAvailable method")
//			},
//			ViewPathFunc: func() string {
//				panic("mock out the ViewPath method")
//			},
//		}
//
//		// use mockedExtension in code that requires Extension
//		// and then make assertions.
//
//	}
type ExtensionMock struct {
	// AliasesFunc mocks the Aliases method.
	AliasesFunc func() ([]Alias, error)

	// CommandsFunc mocks the Commands method.
	CommandsFunc func() ([]Command, error)

	// CurrentVersionFunc mocks the CurrentVersion method.
	CurrentVersionFunc func() string

	// IsBinaryFunc mocks the IsBinary method.
	IsBinaryFunc func() bool

	// IsLocalFunc mocks the IsLocal method.
	IsLocalFunc func() bool

	// IsPinnedFunc mocks the IsPinned method.
	IsPinnedFunc func() bool

	// NameFunc mocks the Name method.
	NameFunc func() string

	// PathFunc mocks the Path method.
	PathFunc func() string

	// TemplatePathFunc mocks the TemplatePath method.
	TemplatePathFunc func() string

	// URLFunc mocks the URL method.
	URLFunc func() string

	// UpdateAvailableFunc mocks the UpdateAvailable method.
	UpdateAvailableFunc func() bool

	// ViewPathFunc mocks the ViewPath method.
	ViewPathFunc func() string

	// calls tracks calls to the methods.
	calls struct {
		// Aliases holds details about calls to the Aliases method.
		Aliases []struct {
		}
		// Commands holds details about calls to the Commands method.
		Commands []struct {
		}
		// CurrentVersion holds details about calls to the CurrentVersion method.
		CurrentVersion []struct {
		}
		// IsBinary holds details about calls to the IsBinary method.
		IsBinary []struct {
		}
		// IsLocal holds details about calls to the IsLocal method.
		IsLocal []struct {
		}
		// IsPinned holds details about calls to the IsPinned method.
		IsPinned []struct {
		}
		// Name holds details about calls to the Name method.
		Name []struct {
		}
		// Path holds details about calls to the Path method.
		Path []struct {
		}
		// TemplatePath holds details about calls to the TemplatePath method.
		TemplatePath []struct {
		}
		// URL holds details about calls to the URL method.
		URL []struct {
		}
		// UpdateAvailable holds details about calls to the UpdateAvailable method.
		UpdateAvailable []struct {
		}
		// ViewPath holds details about calls to the ViewPath method.
		ViewPath []struct {
		}
	}
	lockAliases         sync.RWMutex
	lockCommands        sync.RWMutex
	lockCurrentVersion  sync.RWMutex
	lockIsBinary        sync.RWMutex
	lockIsLocal         sync.RWMutex
	lockIsPinned        sync.RWMutex
	lockName            sync.RWMutex
	lockPath            sync.RWMutex
	lockTemplatePath    sync.RWMutex
	lockURL             sync.RWMutex
	lockUpdateAvailable sync.RWMutex
	lockViewPath        sync.RWMutex
}

// Aliases calls AliasesFunc.
func (mock *ExtensionMock) Aliases() ([]Alias, error) {
	if mock.AliasesFunc == nil {
		panic("ExtensionMock.AliasesFunc: method is nil but Extension.Aliases was just called")
	}
	callInfo := struct {
	}{}
	mock.lockAliases.Lock()
	mock.calls.Aliases = append(mock.calls.Aliases, callInfo)
	mock.lockAliases.Unlock()
	return mock.AliasesFunc()
}

// AliasesCalls gets all the calls that were made to Aliases.
// Check the length with:
//
//	len(mockedExtension.AliasesCalls())
func (mock *ExtensionMock) AliasesCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockAliases.RLock()
	calls = mock.calls.Aliases
	mock.lockAliases.RUnlock()
	return calls
}

// Commands calls CommandsFunc.
func (mock *ExtensionMock) Commands() ([]Command, error) {
	if mock.CommandsFunc == nil {
		panic("ExtensionMock.CommandsFunc: method is nil but Extension.Commands was just called")
	}
	callInfo := struct {
	}{}
	mock.lockCommands.Lock()
	mock.calls.Commands = append(mock.calls.Commands, callInfo)
	mock.lockCommands.Unlock()
	return mock.CommandsFunc()
}

// CommandsCalls gets all the calls that were made to Commands.
// Check the length with:
//
//	len(mockedExtension.CommandsCalls())
func (mock *ExtensionMock) CommandsCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockCommands.RLock()
	calls = mock.calls.Commands
	mock.lockCommands.RUnlock()
	return calls
}

// CurrentVersion calls CurrentVersionFunc.
func (mock *ExtensionMock) CurrentVersion() string {
	if mock.CurrentVersionFunc == nil {
		panic("ExtensionMock.CurrentVersionFunc: method is nil but Extension.CurrentVersion was just called")
	}
	callInfo := struct {
	}{}
	mock.lockCurrentVersion.Lock()
	mock.calls.CurrentVersion = append(mock.calls.CurrentVersion, callInfo)
	mock.lockCurrentVersion.Unlock()
	return mock.CurrentVersionFunc()
}

// CurrentVersionCalls gets all the calls that were made to CurrentVersion.
// Check the length with:
//
//	len(mockedExtension.CurrentVersionCalls())
func (mock *ExtensionMock) CurrentVersionCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockCurrentVersion.RLock()
	calls = mock.calls.CurrentVersion
	mock.lockCurrentVersion.RUnlock()
	return calls
}

// IsBinary calls IsBinaryFunc.
func (mock *ExtensionMock) IsBinary() bool {
	if mock.IsBinaryFunc == nil {
		panic("ExtensionMock.IsBinaryFunc: method is nil but Extension.IsBinary was just called")
	}
	callInfo := struct {
	}{}
	mock.lockIsBinary.Lock()
	mock.calls.IsBinary = append(mock.calls.IsBinary, callInfo)
	mock.lockIsBinary.Unlock()
	return mock.IsBinaryFunc()
}

// IsBinaryCalls gets all the calls that were made to IsBinary.
// Check the length with:
//
//	len(mockedExtension.IsBinaryCalls())
func (mock *ExtensionMock) IsBinaryCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockIsBinary.RLock()
	calls = mock.calls.IsBinary
	mock.lockIsBinary.RUnlock()
	return calls
}

// IsLocal calls IsLocalFunc.
func (mock *ExtensionMock) IsLocal() bool {
	if mock.IsLocalFunc == nil {
		panic("ExtensionMock.IsLocalFunc: method is nil but Extension.IsLocal was just called")
	}
	callInfo := struct {
	}{}
	mock.lockIsLocal.Lock()
	mock.calls.IsLocal = append(mock.calls.IsLocal, callInfo)
	mock.lockIsLocal.Unlock()
	return mock.IsLocalFunc()
}

// IsLocalCalls gets all the calls that were made to IsLocal.
// Check the length with:
//
//	len(mockedExtension.IsLocalCalls())
func (mock *ExtensionMock) IsLocalCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockIsLocal.RLock()
	calls = mock.calls.IsLocal
	mock.lockIsLocal.RUnlock()
	return calls
}

// IsPinned calls IsPinnedFunc.
func (mock *ExtensionMock) IsPinned() bool {
	if mock.IsPinnedFunc == nil {
		panic("ExtensionMock.IsPinnedFunc: method is nil but Extension.IsPinned was just called")
	}
	callInfo := struct {
	}{}
	mock.lockIsPinned.Lock()
	mock.calls.IsPinned = append(mock.calls.IsPinned, callInfo)
	mock.lockIsPinned.Unlock()
	return mock.IsPinnedFunc()
}

// IsPinnedCalls gets all the calls that were made to IsPinned.
// Check the length with:
//
//	len(mockedExtension.IsPinnedCalls())
func (mock *ExtensionMock) IsPinnedCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockIsPinned.RLock()
	calls = mock.calls.IsPinned
	mock.lockIsPinned.RUnlock()
	return calls
}

// Name calls NameFunc.
func (mock *ExtensionMock) Name() string {
	if mock.NameFunc == nil {
		panic("ExtensionMock.NameFunc: method is nil but Extension.Name was just called")
	}
	callInfo := struct {
	}{}
	mock.lockName.Lock()
	mock.calls.Name = append(mock.calls.Name, callInfo)
	mock.lockName.Unlock()
	return mock.NameFunc()
}

// NameCalls gets all the calls that were made to Name.
// Check the length with:
//
//	len(mockedExtension.NameCalls())
func (mock *ExtensionMock) NameCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockName.RLock()
	calls = mock.calls.Name
	mock.lockName.RUnlock()
	return calls
}

// Path calls PathFunc.
func (mock *ExtensionMock) Path() string {
	if mock.PathFunc == nil {
		panic("ExtensionMock.PathFunc: method is nil but Extension.Path was just called")
	}
	callInfo := struct {
	}{}
	mock.lockPath.Lock()
	mock.calls.Path = append(mock.calls.Path, callInfo)
	mock.lockPath.Unlock()
	return mock.PathFunc()
}

// PathCalls gets all the calls that were made to Path.
// Check the length with:
//
//	len(mockedExtension.PathCalls())
func (mock *ExtensionMock) PathCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockPath.RLock()
	calls = mock.calls.Path
	mock.lockPath.RUnlock()
	return calls
}

// TemplatePath calls TemplatePathFunc.
func (mock *ExtensionMock) TemplatePath() string {
	if mock.TemplatePathFunc == nil {
		panic("ExtensionMock.TemplatePathFunc: method is nil but Extension.TemplatePath was just called")
	}
	callInfo := struct {
	}{}
	mock.lockTemplatePath.Lock()
	mock.calls.TemplatePath = append(mock.calls.TemplatePath, callInfo)
	mock.lockTemplatePath.Unlock()
	return mock.TemplatePathFunc()
}

// TemplatePathCalls gets all the calls that were made to TemplatePath.
// Check the length with:
//
//	len(mockedExtension.TemplatePathCalls())
func (mock *ExtensionMock) TemplatePathCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockTemplatePath.RLock()
	calls = mock.calls.TemplatePath
	mock.lockTemplatePath.RUnlock()
	return calls
}

// URL calls URLFunc.
func (mock *ExtensionMock) URL() string {
	if mock.URLFunc == nil {
		panic("ExtensionMock.URLFunc: method is nil but Extension.URL was just called")
	}
	callInfo := struct {
	}{}
	mock.lockURL.Lock()
	mock.calls.URL = append(mock.calls.URL, callInfo)
	mock.lockURL.Unlock()
	return mock.URLFunc()
}

// URLCalls gets all the calls that were made to URL.
// Check the length with:
//
//	len(mockedExtension.URLCalls())
func (mock *ExtensionMock) URLCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockURL.RLock()
	calls = mock.calls.URL
	mock.lockURL.RUnlock()
	return calls
}

// UpdateAvailable calls UpdateAvailableFunc.
func (mock *ExtensionMock) UpdateAvailable() bool {
	if mock.UpdateAvailableFunc == nil {
		panic("ExtensionMock.UpdateAvailableFunc: method is nil but Extension.UpdateAvailable was just called")
	}
	callInfo := struct {
	}{}
	mock.lockUpdateAvailable.Lock()
	mock.calls.UpdateAvailable = append(mock.calls.UpdateAvailable, callInfo)
	mock.lockUpdateAvailable.Unlock()
	return mock.UpdateAvailableFunc()
}

// UpdateAvailableCalls gets all the calls that were made to UpdateAvailable.
// Check the length with:
//
//	len(mockedExtension.UpdateAvailableCalls())
func (mock *ExtensionMock) UpdateAvailableCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockUpdateAvailable.RLock()
	calls = mock.calls.UpdateAvailable
	mock.lockUpdateAvailable.RUnlock()
	return calls
}

// ViewPath calls ViewPathFunc.
func (mock *ExtensionMock) ViewPath() string {
	if mock.ViewPathFunc == nil {
		panic("ExtensionMock.ViewPathFunc: method is nil but Extension.ViewPath was just called")
	}
	callInfo := struct {
	}{}
	mock.lockViewPath.Lock()
	mock.calls.ViewPath = append(mock.calls.ViewPath, callInfo)
	mock.lockViewPath.Unlock()
	return mock.ViewPathFunc()
}

// ViewPathCalls gets all the calls that were made to ViewPath.
// Check the length with:
//
//	len(mockedExtension.ViewPathCalls())
func (mock *ExtensionMock) ViewPathCalls() []struct {
} {
	var calls []struct {
	}
	mock.lockViewPath.RLock()
	calls = mock.calls.ViewPath
	mock.lockViewPath.RUnlock()
	return calls
}
