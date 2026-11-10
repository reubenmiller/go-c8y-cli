package mcpserver

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
	"time"
)

// approvalInputID id of the approval in the input requests of a tool result
const approvalInputID = "approve"

// approvalTTL how long the user has to answer an approval request
const approvalTTL = 15 * time.Minute

var errApprovalMismatch = errors.New("the approval does not match this command or session, or it was already used")

// approvalTokens issues and checks the request state of approval requests. The MCP
// client returns the state with the user's answer (multi round-trip requests,
// SEP-2322), so it is signed and can only be used once, for the same command,
// input and session that the user was asked about.
type approvalTokens struct {
	once    sync.Once
	mu      sync.Mutex
	secret  []byte
	pending map[string]time.Time
}

func (a *approvalTokens) init() {
	a.once.Do(func() {
		a.secret = make([]byte, 32)
		if _, err := rand.Read(a.secret); err != nil {
			panic(err)
		}
		a.pending = map[string]time.Time{}
	})
}

func (a *approvalTokens) sign(nonce string, key string) string {
	mac := hmac.New(sha256.New, a.secret)
	mac.Write([]byte(nonce))
	mac.Write([]byte{0})
	mac.Write([]byte(key))
	return hex.EncodeToString(mac.Sum(nil))
}

// issue returns the request state for an approval of the given command
func (a *approvalTokens) issue(key string) string {
	a.init()
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	nonce := hex.EncodeToString(b)

	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	for n, expires := range a.pending {
		if now.After(expires) {
			delete(a.pending, n)
		}
	}
	a.pending[nonce] = now.Add(approvalTTL)
	return nonce + "." + a.sign(nonce, key)
}

// use checks that the request state belongs to the given command and marks it as used
func (a *approvalTokens) use(state string, key string) error {
	a.init()
	nonce, signature, ok := strings.Cut(state, ".")
	if !ok || !hmac.Equal([]byte(signature), []byte(a.sign(nonce, key))) {
		return errApprovalMismatch
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	expires, found := a.pending[nonce]
	delete(a.pending, nonce)
	if !found || time.Now().After(expires) {
		return errApprovalMismatch
	}
	return nil
}

// approvalKey identifies what the user approves: the command, its input and the session
func approvalKey(session SessionInfo, plan *Plan, in CommandInput) string {
	return strings.Join([]string{plan.String(), in.Stdin, session.Host, session.Tenant, session.Username, session.SessionName}, "\x00")
}
