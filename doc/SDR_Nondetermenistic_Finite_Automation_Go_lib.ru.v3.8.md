# SDR: Nondeterministic Finite Automation library (goNFA)

**Версия:** 3.8
**Дата:** 20.09.2025

## 1. Введение

### 1.1. Назначение библиотеки

`goNFA` — это универсальная, легковесная и идиоматичная библиотека на языке Go для создания и управления недетерминированными конечными автоматами (NFA).

### 1.2. Область применения

Основное применение — предоставление надежного механизма управления состояниями для сложных систем, таких как движки бизнес-процессов (BPM). Библиотека проектируется как универсальное решение, которое может быть использовано в любых других проектах, требующих реализации сложной логики состояний, особенно в долгоживущих процессах.

### 1.3. Терминология

* **Определение (Definition)**: Статическая, неизменяемая структура, описывающая граф состояний, переходов и хуков.

* **Экземпляр (Machine)**: Динамический объект, "живущий" на графе Определения.

* **Состояние Машины (MachineState)**: Интерфейс для read-only доступа к состоянию Экземпляра.

* **Расширитель Состояния (StateExtender)**: Пользовательский бизнес-объект, присоединенный к Экземпляру.

* **Полезная нагрузка (Payload)**: Произвольные данные, специфичные для события, передаваемые в машину во время переходов.

* **Защитник (Guard)**: Объект, **привязанный к конкретному переходу** и отвечающий за проверку условий его выполнения. Цепочка Защитников работает по принципу middleware: переход разрешен только если каждый Защитник в цепочке вернет `true`.

* **Действие (Action)**: Объект, выполняющий полезную работу. **Действия перехода, входа и выхода привязаны к конкретным переходам или состояниям.** Цепочка Действий выполняется последовательно в рамках транзакции перехода.

* **Хук (Hook)**: `Action`, **привязанный ко всей машине**, который вызывается после *любой* попытки перехода (успешной или неуспешной). Цепочка Хуков работает по принципу middleware, выполняясь последовательно для логирования, метрик или других сквозных задач.

* **Реестр (Registry)**: Объект, который сопоставляет строковые имена с реализациями `Guard` и `Action`.

## 2. Высокоуровневая архитектура

Библиотека разделяет статическое **Определение** и его динамические **Экземпляры**. Определение описывает граф состояний и переходы. Экземпляр содержит текущее состояние FSM и несет с собой пользовательский бизнес-объект (`StateExtender`), предоставляя полный контекст для `Guard`'ов и `Action`'ов.

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

// Payload is an interface for passing event-specific runtime data.
type Payload interface{}

// StateExtender is a placeholder for any user-defined business object.
type StateExtender interface{}

// MachineState provides a read-only view of the machine's state.
type MachineState interface {
	CurrentState() State
	History() []HistoryEntry
	IsInFinalState() bool
	// StateExtender returns the attached user-defined business object.
	StateExtender() StateExtender
}

// Guard is the interface for guard objects.
type Guard interface {
	Check(ctx context.Context, state MachineState, payload Payload) bool
}

// Action is the interface for action and hook objects.
type Action interface {
	Execute(ctx context.Context, state MachineState, payload Payload) error
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
// Assuming this code is in package 'definition'

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
	InitialState  State
	FinalStates   map[State]bool // Set of final (accepting) states.
	States        map[State]StateConfig
	Hooks         Hooks
	// internal fields...
}

// Load loads a definition from an io.Reader using a registry and validates it.
func Load(r io.Reader, registry *Registry) (*Definition, error) { /* ... */ }
`

### 3.4. Программный Builder

`
// Assuming this code is in a package like 'builder'

// Builder provides a fluent interface for creating a Definition.
type Builder struct { /* ... */ }

// New creates a new Builder.
func New() *Builder { /* ... */ }

// InitialState sets the initial state for the state machine.
func (b *Builder) InitialState(s State) *Builder { /* ... */ }

// FinalStates sets the final (accepting) states for the state machine.
// Can be called multiple times to add more states.
func (b *Builder) FinalStates(states ...State) *Builder { /* ... */ }

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

// Build finalizes the building process, performs full validation, and returns an immutable Definition.
func (b *Builder) Build() (*Definition, error) { /* ... */ }
`

### 3.5. Экземпляр автомата (Machine)

`
// Assuming this code is in package 'machine' or 'gonfa'

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
// It automatically satisfies the MachineState interface.
type Machine struct {
    // ...
    history       []HistoryEntry
    stateExtender StateExtender
}

// New creates a new instance of the state machine from a definition,
// attaching a user-defined business object as its state extender.
func New(def *Definition, extender StateExtender) *Machine { /* ... */ }

// Restore restores an instance of the state machine from a storable state,
// attaching a user-defined business object as its state extender.
func Restore(def *Definition, state *Storable, extender StateExtender) (*Machine, error) { /* ... */ }

// Fire triggers a transition based on an event. See section 3.7 for execution order and error handling.
func (m *Machine) Fire(ctx context.Context, event Event, payload Payload) (bool, error) { /* ... */ }

// CurrentState returns the current state.
func (m *Machine) CurrentState() State { /* ... */ }

// History returns the transition history.
func (m *Machine) History() []HistoryEntry { return m.history }

// IsInFinalState checks if the machine is currently in a final (accepting) state.
func (m *Machine) IsInFinalState() bool { /* ... */ }

// StateExtender returns the attached user-defined business object.
func (m *Machine) StateExtender() StateExtender { return m.stateExtender }

// Marshal creates a serializable representation of the instance's state.
func (m *Machine) Marshal() (*Storable, error) { /* ... */ }
`

### 3.6. Правила валидации Определения

Функции `builder.Build()` и `definition.Load()` должны выполнять полную проверку корректности и целостности определения перед его созданием. Определение считается валидным, если выполнены все следующие условия:

1. **Начальное состояние определено:** `initialState` не может быть пустым.

2. **Все состояния должны быть явно объявлены:** Любое имя состояния, используемое в `initialState`, `finalStates`, `transitions` (в полях `from` и `to`), должно быть объявлено либо через вызов `OnEntry`/`OnExit` в билдере, либо присутствовать в секции `states` в YAML-файле. Это правило гарантирует отсутствие переходов в "несуществующие" состояния.

3. **Недостижимые состояния:** Опционально, валидатор может выдавать предупреждение (warning), если в определении есть состояния, в которые невозможно попасть из начального состояния (кроме самого начального состояния).

### 3.7. Обработка ошибок и Транзакционность

Переход из одного состояния в другое является **атомарной операцией**. Метод `Fire` гарантирует, что состояние машины останется консистентным.

**Порядок выполнения `Fire`:**

1. Найти все переходы, соответствующие текущему состоянию и событию.

2. Для каждого найденного перехода выполнить проверку `Guard`'ов. Выбирается первый переход, для которого все `Guard`'ы вернули `true`. Если ни один переход не подходит, операция завершается без изменения состояния.

3. **Начало транзакции:**
   a. Выполнить все `Action`'ы из `OnExit` для текущего состояния.
   b. Выполнить все `Action`'ы самого перехода.
   c. Выполнить все `Action`'ы из `OnEntry` для **целевого** состояния.

4. **Завершение транзакции:**

   * **В случае успеха (все `Action`'ы вернули `nil`):**

     * Состояние машины атомарно меняется на целевое.

     * В историю добавляется новая запись.

     * Вызываются хуки `OnSuccess`.

     * Метод возвращает `(true, nil)`.

   * **В случае ошибки (любой `Action` вернул ошибку):**

     * Операция немедленно прерывается.

     * **Состояние машины НЕ меняется.**

     * Вызываются хуки `OnFailure`.

     * Метод возвращает `(false, err)`, где `err` - исходная ошибка от `Action`.

Этот подход гарантирует, что машина не окажется в несогласованном состоянии, и оставляет приложению полный контроль над тем, как реагировать на ошибки выполнения бизнес-логики.

## 4. Рекомендуемый паттерн использования

С новой архитектурой паттерн использования становится значительно чище. `Payload` используется для передачи данных, специфичных для события, а основной бизнес-контекст доступен через `MachineState`.

`
// 1. Ваш бизнес-объект в goBPM
type Instance struct {
    ID string
    Document DocumentData
    State *machine.Storable // Состояние FSM живет внутри вашего объекта
}

// 2. Реализация Guard'а
type ManagerGuard struct {}

func (g *ManagerGuard) Check(ctx context.Context, state machine.MachineState, payload machine.Payload) bool {
    // Получаем основной бизнес-объект через MachineState
    instance, ok := state.StateExtender().(*Instance)
    if !ok { return false /* log error */ }
    
    // Получаем данные события из Payload (если они есть)
    // approvalParams, _ := payload.(ApprovalParams)

    user := user.FromContext(ctx)
    return user.IsManager && user.Department == instance.Document.Department
}

// 3. Использование в сервисе
func (s *Service) ApproveDocument(ctx context.Context, instanceID string, params ApprovalParams) error {
    // Загружаем ваш бизнес-объект
    instance, err := s.repo.GetInstance(instanceID)
    if err != nil { return err }

    // "Оживляем" машину, ПРИСОЕДИНЯЯ к ней ваш бизнес-объект
    m, err := machine.Restore(s.definition, instance.State, instance)
    if err != nil { return err }
    
    // Вызываем Fire, передавая в payload только данные события
    changed, err := m.Fire(ctx, "Approve", params)
    if err != nil { return err }

    if changed {
        storable, _ := m.Marshal()
        instance.State = storable
        return s.repo.SaveInstance(instance)
    }
    
    return nil
}
`

## 5. Нефункциональные требования

* **Производительность**: Минимальные накладные расходы.

* **Надежность**: Полное покрытие тестами (>90%).

* **Документация**: Исчерпывающие godoc-комментарии.

* **Зависимости**: Минимальное количество внешних зависимостей.

## 6. Пример использования

`
// Assuming 'definition' and 'machine' are separate packages

// ... (registering guard/action objects in the Registry) ...

// Loading the definition from a file
file, _ := os.Open("definition.yaml")
def, err := definition.Load(file, registry)
// ...

// Your business object
type MyProcess struct {
    ID string
    Data string
    FSMState *machine.Storable
}
process := &MyProcess{ID: "p1", Data: "initial data"}

// Creating a new machine instance, attaching your business object
m := machine.New(def, process)

// Firing an event with event-specific payload
ctx := context.Background()
eventParams := map[string]string{"user": "Ruslan"}
m.Fire(ctx, "Submit", eventParams)

// Check if the machine has reached an accepting state
if m.IsInFinalState() {
    // ... logic for a completed process
}

// 1. Saving the state
storable, err := m.Marshal()
if err != nil { /* ... */ }
process.FSMState = storable

// ... (save your entire 'process' object to DB)

// ... (later, in another process)

// 2. Restoring the state
// Load your 'process' object from DB first
loadedProcess, err := repo.Get("p1")
if err != nil { /* ... */ }

// Restore the machine instance, re-attaching your business object
restoredMachine, err := machine.Restore(def, loadedProcess.FSMState, loadedProcess)
if err != nil { /* ... */ }
`

## Приложение А: Пример структуры файла определения

`
# The initial state of the machine
initialState: Draft

# The final (accepting) states of the machine
finalStates:
  - Approved
  - Archived

# Global hooks
hooks:
  onSuccess:
    - logSuccess
  onFailure:
    - logFailure

# Description of state-specific actions.
# All states must be listed here, even if they have no actions.
states:
  Draft: {} # Explicitly defined state
  InReview:
    onEntry:
      - assignReviewer
    onExit:
      - cleanupTask
  Approved:
    onEntry:
      - archiveDocument
  Archived: {}
  Rejected: {}

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
    to: In-review
    on: Rework
`

## Приложение Б: JSON Schema для файла определения

**Примечание:** Данная схема валидирует только *структуру* файла. Семантическая валидация (например, проверка того, что все используемые состояния объявлены в секции `states`) выполняется кодом библиотеки при загрузке (см. раздел 3.6).

`
{
  "$schema": "[http://json-schema.org/draft-07/schema#](http://json-schema.org/draft-07/schema#)",
  "$id": "[https://example.com/gonfa-definition.schema.json](https://example.com/gonfa-definition.schema.json)",
  "title": "goNFA Definition",
  "description": "Schema for a goNFA state machine definition file.",
  "type": "object",
  "properties": {
    "initialState": {
      "description": "The initial state of the machine.",
      "type": "string"
    },
    "finalStates": {
      "description": "A list of final (accepting) states.",
      "type": "array",
      "items": { "type": "string" },
      "uniqueItems": true
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
      "description": "State-specific entry and exit actions. All states used in the machine must be defined here.",
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
  "required": ["initialState", "states", "transitions"],
  "additionalProperties": false
}
`