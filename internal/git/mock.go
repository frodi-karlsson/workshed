package git

import (
	"context"
	"sync"
)

type MockGit struct {
	mu                    sync.Mutex
	cloneErr              error
	checkoutErr           error
	getRemoteErr          error
	getRemoteURLResult    string
	currentBranchErr      error
	currentBranchResult   string
	revParseErr           error
	revParseResult        string
	statusPorcelainErr    error
	statusPorcelainResult string
	cloneCalls            []CloneCall
	checkoutCalls         []CheckoutCall
	getRemoteCalls        []GetRemoteCall
	currentBranchCalls    []CurrentBranchCall
	revParseCalls         []RevParseCall
	statusPorcelainCalls  []StatusPorcelainCall
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

type RevParseCall struct {
	Dir string
	Ref string
}

type StatusPorcelainCall struct {
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

func (m *MockGit) RevParse(ctx context.Context, dir, ref string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.revParseCalls = append(m.revParseCalls, RevParseCall{Dir: dir, Ref: ref})
	if m.revParseErr != nil {
		return "", m.revParseErr
	}
	return m.revParseResult, nil
}

func (m *MockGit) SetRevParseErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.revParseErr = err
}

func (m *MockGit) SetRevParseResult(commit string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.revParseResult = commit
}

func (m *MockGit) GetRevParseCalls() []RevParseCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]RevParseCall{}, m.revParseCalls...)
}

func (m *MockGit) StatusPorcelain(ctx context.Context, dir string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.statusPorcelainCalls = append(m.statusPorcelainCalls, StatusPorcelainCall{Dir: dir})
	if m.statusPorcelainErr != nil {
		return "", m.statusPorcelainErr
	}
	return m.statusPorcelainResult, nil
}

func (m *MockGit) SetStatusPorcelainErr(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusPorcelainErr = err
}

func (m *MockGit) SetStatusPorcelainResult(status string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.statusPorcelainResult = status
}

func (m *MockGit) GetStatusPorcelainCalls() []StatusPorcelainCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]StatusPorcelainCall{}, m.statusPorcelainCalls...)
}
