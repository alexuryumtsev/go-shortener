/*
Staticlint - кастомный статический анализатор кода.

Анализатор включает в себя следующие проверки:

Стандартные анализаторы:
  - printf: проверка корректности аргументов в принт-функциях
  - shadow: определение затененных переменных
  - structtag: проверка корректности тегов структур
  - assign: проверка подозрительных присваиваний
  - atomic: проверка корректности использования atomic
  - bools: проверка упрощения булевых операций
  - composites: проверка литералов композитных типов
  - copylocks: проверка некорректного копирования мьютексов

Анализаторы staticcheck:
  - Все анализаторы класса SA
  - ST1000: некорректная документация пакета
  - QF1001: цикл for может быть упрощен до range
  - S1028: упрощение сравнения с пустой строкой

Кастомные анализаторы:
  - errcheck: проверка необработанных ошибок
  - exitcheck: запрет на прямой вызов os.Exit в main

Запуск:

	go run cmd/staticlint/main.go ./...
*/
package main

import (
	"strings"

	"github.com/alexuryumtsev/go-shortener/cmd/staticlint/exitcheckanalyzer"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
	"golang.org/x/tools/go/analysis/passes/asmdecl"
	"golang.org/x/tools/go/analysis/passes/assign"
	"golang.org/x/tools/go/analysis/passes/atomic"
	"golang.org/x/tools/go/analysis/passes/bools"
	"golang.org/x/tools/go/analysis/passes/buildtag"
	"golang.org/x/tools/go/analysis/passes/composite"
	"golang.org/x/tools/go/analysis/passes/copylock"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"honnef.co/go/tools/staticcheck"
)

// Config — имя файла конфигурации
const Config = `config.json`

// ConfigData описывает структуру файла конфигурации
type ConfigData struct {
	Staticcheck []string
}

func main() {
	// Инициализация стандартных анализаторов
	standardAnalyzers := []*analysis.Analyzer{
		printf.Analyzer,
		shadow.Analyzer,
		structtag.Analyzer,
		assign.Analyzer,
		atomic.Analyzer,
		bools.Analyzer,
		composite.Analyzer,
		copylock.Analyzer,
		asmdecl.Analyzer,
		buildtag.Analyzer,
	}

	// Добавляем кастомные анализаторы
	customAnalyzers := []*analysis.Analyzer{
		exitcheckanalyzer.ExitCheckAnalyzer,
	}

	// Подготовка списка анализаторов staticcheck
	var statichecks []*analysis.Analyzer
	for _, v := range staticcheck.Analyzers {
		// Добавляем все SA анализаторы
		if strings.HasPrefix(v.Analyzer.Name, "SA") {
			statichecks = append(statichecks, v.Analyzer)
		}
		// Добавляем выборочные анализаторы из других категорий
		switch v.Analyzer.Name {
		case "ST1000", "QF1001", "S1028":
			statichecks = append(statichecks, v.Analyzer)
		}
	}

	// Объединяем все анализаторы
	allAnalyzers := append(standardAnalyzers, customAnalyzers...)
	allAnalyzers = append(allAnalyzers, statichecks...)

	multichecker.Main(allAnalyzers...)
}
