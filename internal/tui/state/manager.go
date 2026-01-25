package state

import (
	"context"
	"sync"
)

type State string

const (
	StateInit      State = "init"
	StateIdle      State = "idle"
	StateLoading   State = "loading"
	StateActive    State = "active"
	StateError     State = "error"
	StateDone      State = "done"
	StateCancelled State = "cancelled"
)

type StateTransition struct {
	From   State
	To     State
	Action func() error
}

type StateManager struct {
	currentState State
	transitions  map[State][]StateTransition
	listeners    []StateChangeListener
	mu           sync.RWMutex
	history      []State
	maxHistory   int
	data         map[string]interface{}
}

type StateChangeListener struct {
	OnChange func(from, to State)
	ID       string
}

func NewStateManager(initialState State) *StateManager {
	sm := &StateManager{}
	sm.currentState = initialState
	sm.transitions = make(map[State][]StateTransition)
	sm.listeners = []StateChangeListener{}
	sm.history = []State{initialState}
	sm.maxHistory = 10
	sm.data = make(map[string]interface{})
	return sm
}

func (sm *StateManager) GetState() State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.currentState
}

func (sm *StateManager) CanTransition(to State) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	transitions, ok := sm.transitions[sm.currentState]
	if !ok {
		return false
	}

	for _, t := range transitions {
		if t.To == to {
			return true
		}
	}
	return false
}

func (sm *StateManager) Transition(to State) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	transitions, ok := sm.transitions[sm.currentState]
	if !ok {
		return ErrNoTransition
	}

	for _, t := range transitions {
		if t.To == to {
			sm.currentState = to
			sm.addHistory(to)

			sm.mu.Unlock()
			sm.notifyListeners(sm.currentState, to)
			sm.mu.Lock()

			if t.Action != nil {
				return t.Action()
			}
			return nil
		}
	}

	return ErrNoTransition
}

func (sm *StateManager) TransitionWithAction(to State, action func() error) error {
	sm.mu.Lock()
	oldState := sm.currentState
	sm.currentState = to
	sm.addHistory(to)
	sm.mu.Unlock()

	sm.notifyListeners(oldState, to)

	if action != nil {
		return action()
	}
	return nil
}

func (sm *StateManager) AddTransition(from, to State, action func() error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.transitions[from] = append(sm.transitions[from], StateTransition{
		From:   from,
		To:     to,
		Action: action,
	})
}

func (sm *StateManager) AddTransitionWithCallback(from, to State, action func() error) {
	sm.AddTransition(from, to, action)
}

func (sm *StateManager) AddListener(listener StateChangeListener) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.listeners = append(sm.listeners, listener)
}

func (sm *StateManager) RemoveListener(id string) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for i, l := range sm.listeners {
		if l.ID == id {
			sm.listeners = append(sm.listeners[:i], sm.listeners[i+1:]...)
			return
		}
	}
}

func (sm *StateManager) notifyListeners(from, to State) {
	for _, l := range sm.listeners {
		if l.OnChange != nil {
			go l.OnChange(from, to)
		}
	}
}

func (sm *StateManager) addHistory(state State) {
	sm.history = append(sm.history, state)
	if len(sm.history) > sm.maxHistory {
		sm.history = sm.history[len(sm.history)-sm.maxHistory:]
	}
}

func (sm *StateManager) GetHistory() []State {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make([]State, len(sm.history))
	copy(result, sm.history)
	return result
}

func (sm *StateManager) CanGoBack() bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.history) > 1
}

func (sm *StateManager) GoBack() (State, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if len(sm.history) <= 1 {
		return sm.currentState, ErrNoHistory
	}

	sm.history = sm.history[:len(sm.history)-1]
	newState := sm.history[len(sm.history)-1]
	oldState := sm.currentState
	sm.currentState = newState

	sm.mu.Unlock()
	sm.notifyListeners(oldState, newState)
	sm.mu.Lock()

	return newState, nil
}

func (sm *StateManager) SetData(key string, value interface{}) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.data[key] = value
}

func (sm *StateManager) GetData(key string) (interface{}, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	v, ok := sm.data[key]
	return v, ok
}

func (sm *StateManager) ClearData() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	sm.data = make(map[string]interface{})
}

func (sm *StateManager) Reset(newState State) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	oldState := sm.currentState
	sm.currentState = newState
	sm.history = []State{newState}
	sm.data = make(map[string]interface{})

	sm.mu.Unlock()
	sm.notifyListeners(oldState, newState)
	sm.mu.Lock()
}

type ModalStack struct {
	modals    []ModalInfo
	stateMgr  *StateManager
	maxModals int
}

type ModalInfo struct {
	ID      string
	Name    string
	State   State
	Context context.Context
	Cancel  context.CancelFunc
	Data    map[string]interface{}
	OnOpen  func() error
	OnClose func() error
}

func NewModalStack(stateManager *StateManager) *ModalStack {
	ms := &ModalStack{}
	ms.stateMgr = stateManager
	ms.modals = []ModalInfo{}
	ms.maxModals = 5
	return ms
}

func (ms *ModalStack) Open(id, name string, ctx context.Context) error {
	if len(ms.modals) >= ms.maxModals {
		return ErrMaxModals
	}

	for _, m := range ms.modals {
		if m.ID == id {
			return ErrModalAlreadyOpen
		}
	}

	childCtx, cancel := context.WithCancel(ctx)

	modal := ModalInfo{
		ID:      id,
		Name:    name,
		State:   StateInit,
		Context: childCtx,
		Cancel:  cancel,
		Data:    make(map[string]interface{}),
	}

	if modal.OnOpen != nil {
		if err := modal.OnOpen(); err != nil {
			cancel()
			return err
		}
	}

	ms.modals = append(ms.modals, modal)

	return nil
}

func (ms *ModalStack) Close(id string) error {
	for i, m := range ms.modals {
		if m.ID == id {
			if m.OnClose != nil {
				if err := m.OnClose(); err != nil {
					return err
				}
			}

			m.Cancel()
			ms.modals = append(ms.modals[:i], ms.modals[i+1:]...)
			return nil
		}
	}

	return ErrModalNotFound
}

func (ms *ModalStack) CloseTop() error {
	if len(ms.modals) == 0 {
		return ErrNoModals
	}

	m := ms.modals[len(ms.modals)-1]
	return ms.Close(m.ID)
}

func (ms *ModalStack) CloseAll() error {
	var errs []error

	for _, m := range ms.modals {
		if err := ms.Close(m.ID); err != nil {
			errs = append(errs, err)
		}
	}

	ms.modals = []ModalInfo{}

	if len(errs) > 0 {
		return errs[0]
	}
	return nil
}

func (ms *ModalStack) GetTop() *ModalInfo {
	if len(ms.modals) == 0 {
		return nil
	}
	return &ms.modals[len(ms.modals)-1]
}

func (ms *ModalStack) GetByID(id string) *ModalInfo {
	for _, m := range ms.modals {
		if m.ID == id {
			return &m
		}
	}
	return nil
}

func (ms *ModalStack) GetAll() []ModalInfo {
	result := make([]ModalInfo, len(ms.modals))
	copy(result, ms.modals)
	return result
}

func (ms *ModalStack) Len() int {
	return len(ms.modals)
}

func (ms *ModalStack) IsEmpty() bool {
	return len(ms.modals) == 0
}

func (ms *ModalStack) IsModalOpen(id string) bool {
	for _, m := range ms.modals {
		if m.ID == id {
			return true
		}
	}
	return false
}

type ModalStackError struct {
	Message string
}

func (e ModalStackError) Error() string {
	return e.Message
}

var (
	ErrNoTransition     = ModalStackError{Message: "no valid transition"}
	ErrNoHistory        = ModalStackError{Message: "no history available"}
	ErrMaxModals        = ModalStackError{Message: "maximum modals reached"}
	ErrModalAlreadyOpen = ModalStackError{Message: "modal already open"}
	ErrModalNotFound    = ModalStackError{Message: "modal not found"}
	ErrNoModals         = ModalStackError{Message: "no modals open"}
)
