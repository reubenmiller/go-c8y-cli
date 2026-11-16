package extensions

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// ErrUnknownSessionProvider is returned when no installed extension provides the session provider
var ErrUnknownSessionProvider = errors.New("unknown session provider")

// ErrAmbiguousSessionProvider is returned when a short provider name is provided by more than one extension
var ErrAmbiguousSessionProvider = errors.New("ambiguous session provider")

// InstalledSessionProvider is a session provider of an installed extension
type InstalledSessionProvider struct {
	Extension Extension
	Provider  SessionProvider
}

// Ref returns the qualified reference of the provider: <extension>/<provider>
func (p InstalledSessionProvider) Ref() string {
	return p.Extension.Name() + "/" + p.Provider.Name
}

// ListSessionProviders returns the session providers of the given extensions, sorted by their
// qualified reference. warn is called for each ignored provider (it can be nil)
func ListSessionProviders(exts []Extension, warn func(ext Extension, msg string)) []InstalledSessionProvider {
	providers := []InstalledSessionProvider{}
	for _, ext := range exts {
		manifest, err := ext.Manifest()
		if err != nil {
			if warn != nil {
				warn(ext, fmt.Sprintf("could not read the extension manifest. %s", err))
			}
			continue
		}
		valid, warnings := manifest.ValidSessionProviders()
		if warn != nil {
			for _, w := range warnings {
				warn(ext, w)
			}
		}
		for _, provider := range valid {
			providers = append(providers, InstalledSessionProvider{Extension: ext, Provider: provider})
		}
	}
	sort.SliceStable(providers, func(i, j int) bool {
		return providers[i].Ref() < providers[j].Ref()
	})
	return providers
}

func providerRefs(providers []InstalledSessionProvider) []string {
	refs := make([]string, 0, len(providers))
	for _, p := range providers {
		refs = append(refs, p.Ref())
	}
	return refs
}

// ResolveSessionProvider finds a session provider by its reference, either <provider> (if only
// one extension provides a provider with that name) or <extension>/<provider>
func ResolveSessionProvider(providers []InstalledSessionProvider, ref string) (*InstalledSessionProvider, error) {
	ref = strings.TrimSpace(ref)
	available := strings.Join(providerRefs(providers), ", ")
	if available == "" {
		available = "<none>"
	}

	if extName, name, qualified := strings.Cut(ref, "/"); qualified {
		extName = strings.TrimPrefix(extName, "c8y-")
		for i, p := range providers {
			if p.Extension.Name() == extName && p.Provider.Name == name {
				return &providers[i], nil
			}
		}
		return nil, fmt.Errorf("%w %q. available=%s", ErrUnknownSessionProvider, ref, available)
	}

	matches := []InstalledSessionProvider{}
	for _, p := range providers {
		if p.Provider.Name == ref {
			matches = append(matches, p)
		}
	}
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("%w %q. available=%s", ErrUnknownSessionProvider, ref, available)
	case 1:
		return &matches[0], nil
	}
	return nil, fmt.Errorf("%w %q, it is provided by more than one extension. Use one of: %s", ErrAmbiguousSessionProvider, ref, strings.Join(providerRefs(matches), ", "))
}

// SessionProviderCompletions returns the completion values for the providers, in the format
// "<ref>\t<description>". The short name is used when it is unique, otherwise the qualified reference
func SessionProviderCompletions(providers []InstalledSessionProvider) []string {
	counts := map[string]int{}
	for _, p := range providers {
		counts[p.Provider.Name]++
	}
	values := make([]string, 0, len(providers))
	for _, p := range providers {
		ref := p.Provider.Name
		if counts[ref] > 1 {
			ref = p.Ref()
		}
		description := p.Provider.Description
		if description == "" {
			description = "Session provider from extension " + p.Extension.Name()
		}
		values = append(values, ref+"\t"+description)
	}
	return values
}
