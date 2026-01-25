package git

import (
	"context"
	"sync"
)

type MockGit struct {
	mu                  sync.Mutex
	cloneErr            error
	checkoutErr         error
	getRemoteErr        error
	getRemoteURLResult  string
	currentBranchErr    error
	currentBranchResult string
	cloneCalls          []CloneCall
	checkoutCalls       []CheckoutCall
	getRemoteCalls      []GetRemoteCall
	currentBranchCalls  []CurrentBranchCall
}

type CloneCall struct {
	URL  string
	Dir  string
	Opts CloneOptions
}

type CheckoutCall struct {
	Dir string
	Ref string
}

type GetRemoteCall struct {
	Dir string
}

type CurrentBranchCall struct {
	Dir string
}

func (m *MockGit) Clone(ctx context.Context, url, dir string, opts CloneOptions) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.cloneCalls = append(m.cloneCalls, CloneCall{URL: url, Dir: dir, Opts: opts})
	return m.cloneErr
}

func (m *MockGit) Checkout(ctx context.Context, dir, ref string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.checkoutCalls = append(m.checkoutCalls, CheckoutCall{Dir: dir, Ref: ref})
	return m.checkoutErr
}

func (m *MockGit) GetRemoteURL(ctx context.Context, dir string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.getRemoteCalls = append(m.getRemoteCalls, GetRemoteCall{Dir: dir})
	if m.getRemoteErr != nil {
		return "", m.getRemoteErr
	}
	return m.getRemoteURLResult, nil
}

func (m *MockGit) SetCloneErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cloneErr = err
}

func (m *MockGit) SetCheckoutErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.checkoutErr = err
}

func (m *MockGit) SetGetRemoteErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getRemoteErr = err
}

func (m *MockGit) SetGetRemoteURLResult(url string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.getRemoteURLResult = url
}

func (m *MockGit) GetCloneCalls() []CloneCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]CloneCall{}, m.cloneCalls...)
}

func (m *MockGit) GetCheckoutCalls() []CheckoutCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]CheckoutCall{}, m.checkoutCalls...)
}

func (m *MockGit) GetGetRemoteCalls() []GetRemoteCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]GetRemoteCall{}, m.getRemoteCalls...)
}

func (m *MockGit) CurrentBranch(ctx context.Context, dir string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.currentBranchCalls = append(m.currentBranchCalls, CurrentBranchCall{Dir: dir})
	if m.currentBranchErr != nil {
		return "", m.currentBranchErr
	}
	return m.currentBranchResult, nil
}

func (m *MockGit) SetCurrentBranchErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentBranchErr = err
}

func (m *MockGit) SetCurrentBranchResult(branch string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.currentBranchResult = branch
}

func (m *MockGit) GetCurrentBranchCalls() []CurrentBranchCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]CurrentBranchCall{}, m.currentBranchCalls...)
}
