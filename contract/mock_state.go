// mock_state.go
package main

type MockState struct {
	db map[string]string
}

func NewMockState() *MockState {
	return &MockState{db: make(map[string]string)}
}

func (m *MockState) Set(key, value string) {
	m.db[key] = value
}

func (m *MockState) Get(key string) *string {
	val, ok := m.db[key]
	if !ok {
		return nil
	}
	return &val
}

func (m *MockState) Delete(key string) {
	delete(m.db, key)
}
