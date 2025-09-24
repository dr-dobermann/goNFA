# goNFA - Библиотека недетерминированных конечных автоматов для Go

[![Go Version](https://img.shields.io/github/go-mod/go-version/dr-dobermann/gonfa)](https://golang.org/)
[![GitHub release](https://img.shields.io/github/v/release/dr-dobermann/gonfa)](https://github.com/dr-dobermann/gonfa/releases)
[![License](https://img.shields.io/badge/License-LGPL%202.1-green.svg)](LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/dr-dobermann/gonfa)](https://goreportcard.com/report/github.com/dr-dobermann/gonfa)
[![codecov](https://codecov.io/gh/dr-dobermann/gonfa/branch/master/graph/badge.svg)](https://codecov.io/gh/dr-dobermann/gonfa)
[![CI/CD Pipeline](https://github.com/dr-dobermann/gonfa/workflows/CI/CD%20Pipeline/badge.svg)](https://github.com/dr-dobermann/gonfa/actions)
[![GoDoc](https://godoc.org/github.com/dr-dobermann/gonfa?status.svg)](https://godoc.org/github.com/dr-dobermann/gonfa)
[![GitHub issues](https://img.shields.io/github/issues/dr-dobermann/gonfa)](https://github.com/dr-dobermann/gonfa/issues)
[![GitHub stars](https://img.shields.io/github/stars/dr-dobermann/gonfa)](https://github.com/dr-dobermann/gonfa/stargazers)

Универсальная, легковесная и идиоматичная библиотека на языке Go для создания и управления недетерминированными конечными автоматами (NFA). goNFA предоставляет надежные механизмы управления состояниями для сложных систем, таких как движки бизнес-процессов (BPM), системы workflow и любые приложения, требующие сложной логики конечных автоматов.

## Возможности

- **Поддержка недетерминированных конечных автоматов**: Полная реализация NFA с множественными переходами из одного состояния по одному событию
- **Потокобезопасные операции**: Все операции машины безопасны для конкурентного доступа
- **Fluent Builder API**: Интуитивное программное создание конечных автоматов
- **YAML конфигурация**: Загрузка определений конечных автоматов из YAML файлов
- **Персистентность состояния**: Сериализация и восстановление состояния машины для долгоживущих процессов
- **Интеграция бизнес-объектов**: Прикрепление пользовательских бизнес-объектов как StateExtender
- **Поддержка финальных состояний**: Явная поддержка принимающих/финальных состояний
- **Расширяемые Actions и Guards**: Плагинная система для пользовательской бизнес-логики с полным доступом к контексту
- **Комплексное тестирование**: >90% покрытие кода с обширными unit и интеграционными тестами
- **Нулевые внешние зависимости**: Основная библиотека не имеет внешних зависимостей (кроме поддержки YAML)

## Быстрый старт

### Установка

```bash
go get github.com/dr-dobermann/gonfa
```

### Базовое использование

```go
package main

import (
    "context"
    "fmt"
    "log"

    "github.com/dr-dobermann/gonfa/pkg/builder"
    "github.com/dr-dobermann/gonfa/pkg/gonfa"
    "github.com/dr-dobermann/gonfa/pkg/machine"
)

// Document представляет ваш бизнес-объект
type Document struct {
    ID     string
    Title  string
    Author string
}

// Простая реализация guard
type ManagerGuard struct{}

func (g *ManagerGuard) Check(ctx context.Context, state gonfa.MachineState, payload gonfa.Payload) bool {
    // Доступ к бизнес-объекту через StateExtender
    doc := state.StateExtender().(*Document)
    fmt.Printf("Проверка одобрения для документа: %s\n", doc.Title)
    return true
}

// Простая реализация action
type NotifyAction struct{}

func (a *NotifyAction) Execute(ctx context.Context, state gonfa.MachineState, payload gonfa.Payload) error {
    // Доступ к бизнес-объекту через StateExtender
    doc := state.StateExtender().(*Document)
    fmt.Printf("Уведомление о документе: %s\n", doc.Title)
    return nil
}

func main() {
    // Создание определения конечного автомата
    definition, err := builder.New().
        InitialState("Черновик").
        FinalStates("Одобрено").
        AddTransition("Черновик", "НаРассмотрении", "Отправить").
        WithActions(&NotifyAction{}).
        AddTransition("НаРассмотрении", "Одобрено", "Одобрить").
        WithGuards(&ManagerGuard{}).
        Build()
    if err != nil {
        log.Fatal(err)
    }

    // Создание бизнес-объекта
    doc := &Document{
        ID:     "DOC-001",
        Title:  "Предложение проекта",
        Author: "Алиса Смит",
    }

    // Создание экземпляра машины с бизнес-объектом
    sm, err := machine.New(definition, doc)
    if err != nil {
        log.Fatal(err)
    }
    
    // Запуск событий
    ctx := context.Background()
    success, err := sm.Fire(ctx, "Отправить", nil)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Переход успешен: %v\n", success)
    fmt.Printf("Текущее состояние: %s\n", sm.CurrentState())
    fmt.Printf("Финальное состояние: %v\n", sm.IsInFinalState())
}
```

### YAML конфигурация

Создание конечного автомата из YAML конфигурации:

```yaml
# workflow.yaml
initialState: Черновик

hooks:
  onSuccess:
    - logSuccess
  onFailure:
    - logFailure

states:
  НаРассмотрении:
    onEntry:
      - assignReviewer
    onExit:
      - cleanupTask

transitions:
  - from: Черновик
    to: НаРассмотрении
    on: Отправить
    actions:
      - notifyAuthor
  - from: НаРассмотрении
    to: Одобрено
    on: Одобрить
    guards: 
      - isManager
```

```go
// Загрузка из YAML
file, err := os.Open("workflow.yaml")
if err != nil {
    log.Fatal(err)
}
defer file.Close()

registry := registry.New()
// Регистрируем ваши guards и actions...
registry.RegisterGuard("isManager", &ManagerGuard{})
registry.RegisterAction("notifyAuthor", &NotifyAction{})

definition, err := definition.LoadDefinition(file, registry)
if err != nil {
    log.Fatal(err)
}

machine := machine.NewMachine(definition)
```

## Архитектура

goNFA разделяет статические **Определения** и динамические **Экземпляры машин**:

- **Definition**: Неизменяемое описание графа состояний, переходов и связанных действий
- **Machine**: Runtime экземпляр, который "живет" на графе Definition с текущим состоянием
- **Registry**: Отображает строковые имена на реализации Guard/Action для загрузки YAML
- **Builder**: Fluent API для программного создания Definition

## Структура пакетов

- [`pkg/gonfa`](pkg/gonfa/README.md) - Основные типы и интерфейсы
- [`pkg/definition`](pkg/definition/README.md) - Определения конечных автоматов и загрузка YAML
- [`pkg/builder`](pkg/builder/README.md) - Fluent API для создания определений
- [`pkg/machine`](pkg/machine/README.md) - Runtime реализация конечных автоматов
- [`pkg/registry`](pkg/registry/README.md) - Отображение имен на объекты для поддержки YAML
- [`examples/`](examples/) - Примеры использования и конфигурации

## Документация

- [Техническая спецификация](doc/SDR_Nondetermenistic_Finite_Automation_Go_lib.ru.md) - Подробные технические требования
- [Техническая спецификация (EN)](doc/SDR_Nondetermenistic_Finite_Automation_Go_lib.en.md) - English version
- [Документация API](https://pkg.go.dev/github.com/dr-dobermann/gonfa) - Справочник GoDoc
- [Примеры](examples/) - Рабочие примеры кода
- [История изменений](CHANGELOG.md) - История версий и изменения
- [English README](README.md) - English version

## Сборка и тестирование

### Предварительные требования

- Go 1.21 или новее
- Make (опционально, для удобства команд)

### Команды разработки

```bash
# Установка инструментов разработки
make install

# Запуск тестов с покрытием
make test

# Сборка библиотеки
make build

# Сборка примеров
make examples

# Запуск конкретного примера
make run-example-document_workflow

# Генерация моков (требует mockery)
make mocks

# Проверка кода линтером
make lint

# Цикл разработки (очистка, моки, тесты)
make dev
```

### Ручные команды

```bash
# Запуск тестов
go test ./pkg/...

# Сборка примеров
go build -o bin/document_workflow examples/document_workflow.go

# Запуск примера
./bin/document_workflow
```

## Вклад в проект

1. Сделайте fork репозитория
2. Создайте ветку для функции (`git checkout -b feature/amazing-feature`)
3. Внесите изменения
4. Добавьте тесты для новой функциональности
5. Убедитесь, что все тесты проходят (`make test`)
6. Зафиксируйте изменения (`git commit -m 'Add amazing feature'`)
7. Отправьте в ветку (`git push origin feature/amazing-feature`)
8. Откройте Pull Request

### Стандарты кода

- Следуйте соглашениям и идиомам Go
- Поддерживайте покрытие тестами >90%
- Добавляйте подробную документацию для публичных API
- Используйте осмысленные сообщения коммитов
- По возможности держите длину строк ≤80 символов

## Лицензия

Этот проект лицензирован под GNU Lesser General Public License v2.1 - смотрите файл [LICENSE](LICENSE) для подробностей.

## Автор

**dr-dobermann** (rgabtiov@gmail.com)

## Ссылки

- [Репозиторий GitHub](https://github.com/dr-dobermann/gonfa)
- [Трекер проблем](https://github.com/dr-dobermann/gonfa/issues)
- [Обсуждения](https://github.com/dr-dobermann/gonfa/discussions)

---

*goNFA - Привносим мощь недетерминированных конечных автоматов в Go приложения.*
