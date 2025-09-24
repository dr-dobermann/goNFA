# SDR: Nondeterministic Finite Automation Go library (goNFA)

**Версия:** 2.6
**Дата:** 20.09.2025

## 1. Введение

### 1.1. Назначение библиотеки

`goNFA` — это универсальная, легковесная и идиоматичная библиотека на языке Go для создания и управления недетерминированными конечными автоматами (NFA).

### 1.2. Область применения

Основное применение — предоставление надежного механизма управления состояниями для сложных систем, таких как движки бизнес-процессов (BPM). Библиотека проектируется как универсальное решение, которое может быть использовано в любых других проектах, требующих реализации сложной логики состояний, особенно в долгоживущих процессах.

### 1.3. Терминология

* **Определение (Definition)**: Статическая, неизменяемая структура, описывающая граф состояний, переходов и хуков.
* **Экземпляр (Machine)**: Динамический объект, "живущий" на графе Определения, имеющий текущее состояние.
* **Полезная нагрузка (Payload)**: Произвольные данные, передаваемые в машину во время переходов.
* **Защитник (Guard)**: Объект, реализующий интерфейс `Guard`, разрешающий или запрещающий переход.
* **Действие (Action)**: Общий термин для объектов, реализующих интерфейс `Action`.
  * **Действие перехода**: Выполняется во время перехода.
  * **Действие на вход/выход**: Выполняется при входе в состояние или выходе из него.
* **Хук (Hook)**: Объект, реализующий интерфейс `Action`, который вызывается после попытки перехода.
* **Реестр (Registry)**: Объект, который сопоставляет строковые имена с реализациями `Guard` и `Action`.

## 2. Высокоуровневая архитектура

Библиотека разделяет статическое **Определение** и его динамические **Экземпляры**. Определение описывает граф состояний, переходы между ними, а также действия, связанные с переходами и самими состояниями. Для связи декларативных определений (из файлов) с кодом используется **Реестр** объектов.

* **Определение** создается программно с помощью **Builder**'а или загружается из файла.
* **Экземпляр** создается из Определения, работает в рантайме и может быть сохранен и восстановлен.
* Все операции над Экземпляром потокобезопасны.

## 3. Проектирование API и структур данных

### 3.1. Основные типы и интерфейсы

`
import (
	"context"
	"io"
	"time"
)

// State represents a state in the state machine.
type State string

// Event represents an event that triggers a transition.
type Event string

// Payload is an interface for passing runtime data.
type Payload interface{}

// Guard is the interface for guard objects.
type Guard interface {
	Check(ctx context.Context, payload Payload) bool
}

// Action is the interface for action and hook objects.
type Action interface {
	Execute(ctx context.Context, payload Payload) error
}
`

### 3.2. Реестр объектов (Registry)

`
// Assuming this code is in package 'registry'

// Registry stores a mapping from string names to real objects.
type Registry struct { /* ... */ }

// New creates a new Registry.
func New() *Registry { /* ... */ }

// RegisterGuard registers a guard object under a unique name.
func (r *Registry) RegisterGuard(name string, guard Guard) error { /* ... */ }

// RegisterAction registers an action (or hook) object under a unique name.
func (r *Registry) RegisterAction(name string, action Action) error { /* ... */ }
`

### 3.3. Определение автомата (Definition)

`
// Transition describes one possible transition.
type Transition struct {
	From    State
	To      State
	On      Event
	Guards  []Guard  // A chain of guards.
	Actions []Action // A chain of transition actions.
}

// StateConfig describes actions associated with a specific state.
type StateConfig struct {
	OnEntry []Action // Actions to execute upon entering the state.
	OnExit  []Action // Actions to execute upon exiting the state.
}

// Hooks describes a set of hooks for the state machine.
type Hooks struct {
	OnSuccess []Action
	OnFailure []Action
}

// Definition is an immutable description of the state graph.
type Definition struct { 
	// ...
	States map[State]StateConfig
	Hooks  Hooks
}

// LoadDefinition loads a definition from an io.Reader using a registry.
func LoadDefinition(r io.Reader, registry *Registry) (*Definition, error) { /* ... */ }

// NewMachine creates a new instance of the state machine.
func (d *Definition) NewMachine() *Machine { /* ... */ }

// RestoreMachine restores an instance of the state machine.
func (d *Definition) RestoreMachine(state *Storable) (*Machine, error) { /* ... */ }
`

### 3.4. Программный Builder

`
// Assuming this code is in a package like 'gonfa' or 'builder'

// Builder provides a fluent interface for creating a Definition.
type Builder struct { /* ... */ }

// New creates a new Builder.
func New() *Builder { /* ... */ }

// InitialState sets the initial state for the state machine.
func (b *Builder) InitialState(s State) *Builder { /* ... */ }

// OnEntry defines actions to be executed upon EVERY entry into the specified state.
func (b *Builder) OnEntry(s State, actions ...Action) *Builder { /* ... */ }

// OnExit defines actions to be executed upon EVERY exit from the specified state.
func (b *Builder) OnExit(s State, actions ...Action) *Builder { /* ... */ }

// AddTransition adds a new transition.
func (b *Builder) AddTransition(from State, to State, on Event) *Builder { /* ... */ }

// WithGuards adds guards to the LAST added transition.
func (b *Builder) WithGuards(guards ...Guard) *Builder { /* ... */ }

// WithActions adds actions to the LAST added transition.
func (b *Builder) WithActions(actions ...Action) *Builder { /* ... */ }

// WithHooks sets global hooks for the state machine.
func (b *Builder) WithHooks(hooks Hooks) *Builder { /* ... */ }

// Build finalizes the building process and returns an immutable Definition.
func (b *Builder) Build() (*Definition, error) { /* ... */ }
`

### 3.5. Экземпляр автомата (Machine) и его состояние

`
// HistoryEntry is for recording transition history.
type HistoryEntry struct {
	From      State
	To        State
	On        Event
	Timestamp time.Time
}

// Storable represents a serializable state of a Machine instance.
type Storable struct {
	CurrentState State          `json:"currentState"`
	History      []HistoryEntry `json:"history"`
}

// Machine represents an instance of a state machine.
type Machine struct { /* ... */ }

// Fire triggers a transition based on an event.
// The provided context is passed down to all Guards and Actions.
// Execution order:
// 1. Check Guards
// 2. Execute OnExit for the current state
// 3. Execute transition Actions
// 4. Change state
// 5. Execute OnEntry for the new state
// 6. Call Hooks (OnSuccess/OnFailure)
func (m *Machine) Fire(ctx context.Context, event Event, payload Payload) (bool, error) { /* ... */ }

// CurrentState returns the current state.
func (m *Machine) CurrentState() State { /* ... */ }

// Marshal creates a serializable representation of the instance's state.
func (m *Machine) Marshal() (*Storable, error) { /* ... */ }
`

## 4. Нефункциональные требования

* **Производительность**: Минимальные накладные расходы.
* **Надежность**: Полное покрытие тестами (>90%).
* **Документация**: Исчерпывающие godoc-комментарии.
* **Зависимости**: Минимальное количество внешних зависимостей.

## 5. Пример использования

`
// ... (registering guard/action objects in the Registry) ...

// Loading the definition from a file
file, _ := os.Open("definition.yaml")
definition, err := LoadDefinition(file, registry)
// ...

// Creating an instance
machine := definition.NewMachine()

// Firing an event with a context and a specific payload
ctx := context.Background() // Or a context from an HTTP request
submissionData := map[string]string{"author": "Ruslan", "file": "doc.pdf"}
machine.Fire(ctx, "Submit", submissionData)

// 1. Saving the state
storable, err := machine.Marshal()
if err != nil { /* ... */ }

// ... (serializing storable to JSON and saving to the DB)
savedData, err := json.Marshal(storable)

// ... (later, in another process)

// 2. Restoring the state
var loadedState Storable
err = json.Unmarshal(savedData, &loadedState)
// ...

// Restoring the instance itself using the same Definition
restoredMachine, err := definition.RestoreMachine(&loadedState)
if err != nil { /* ... */ }
`

## Приложение А: Пример структуры файла определения

`
# The initial state of the machine
initialState: Draft

# Global hooks
hooks:
  onSuccess:
    - logSuccess
  onFailure:
    - logFailure

# Description of state-specific actions
states:
  InReview:
    onEntry:
      - assignReviewer
    onExit:
      - cleanupTask
  Approved:
    onEntry:
      - archiveDocument

# List of all possible transitions
transitions:
  - from: Draft
    to: InReview
    on: Submit
    actions:
      - notifyAuthor

  - from: InReview
    to: Approved
    on: Approve
    guards: 
      - isManager

  - from: Rejected
    to: InReview
    on: Rework
`

## Приложение Б: JSON Schema для файла определения

`
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "https://example.com/gonfa-definition.schema.json",
  "title": "goNFA Definition",
  "description": "Schema for a goNFA state machine definition file.",
  "type": "object",
  "properties": {
    "initialState": {
      "description": "The initial state of the machine.",
      "type": "string"
    },
    "hooks": {
      "description": "Global hooks for all transitions.",
      "type": "object",
      "properties": {
        "onSuccess": {
          "type": "array",
          "items": { "type": "string" },
          "uniqueItems": true
        },
        "onFailure": {
          "type": "array",
          "items": { "type": "string" },
          "uniqueItems": true
        }
      },
      "additionalProperties": false
    },
    "states": {
      "description": "State-specific entry and exit actions.",
      "type": "object",
      "patternProperties": {
        "^.+$": {
          "type": "object",
          "properties": {
            "onEntry": {
              "type": "array",
              "items": { "type": "string" },
              "uniqueItems": true
            },
            "onExit": {
              "type": "array",
              "items": { "type": "string" },
              "uniqueItems": true
            }
          },
          "additionalProperties": false
        }
      }
    },
    "transitions": {
      "description": "The list of all possible transitions.",
      "type": "array",
      "items": {
        "type": "object",
        "properties": {
          "from": { "type": "string" },
          "to": { "type": "string" },
          "on": { "type": "string" },
          "guards": {
            "type": "array",
            "items": { "type": "string" },
            "uniqueItems": true
          },
          "actions": {
            "type": "array",
            "items": { "type": "string" },
            "uniqueItems": true
          }
        },
        "required": ["from", "to", "on"],
        "additionalProperties": false
      }
    }
  },
  "required": ["initialState", "transitions"],
  "additionalProperties": false
}
`
