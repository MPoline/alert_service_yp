// Package staticlint предоставляет набор статических анализаторов для проверки кода на Go.
// Включает как стандартные анализаторы из golang.org/x/tools, так и кастомные проверки.
//
// # Использование
//
// 1. Импортируйте пакет и используйте GetAnalyzers() для получения всех анализаторов:
//
//	import "yourmodule/cmd/staticlint"
//
//	multichecker.Main(staticlint.GetAnalyzers()...)
//
// 2. Или используйте отдельные анализаторы:
//
//	multichecker.Main(staticlint.NoOsExitAnalyzer)
//
// # Доступные анализаторы
//
// Пакет включает следующие анализаторы:
package staticlint

import (
	"go/ast"
	"strings"

	"github.com/fatih/errwrap/errwrap"
	"github.com/masibw/goone"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/nilfunc"
	"golang.org/x/tools/go/analysis/passes/printf"
	"golang.org/x/tools/go/analysis/passes/shadow"
	"golang.org/x/tools/go/analysis/passes/stdmethods"
	"golang.org/x/tools/go/analysis/passes/structtag"
	"golang.org/x/tools/go/analysis/passes/unmarshal"
)

// NoOsExitAnalyzer проверяет запрет на вызовы os.Exit в функции main.
//
// Назначение:
// - Обнаруживает и запрещает прямые вызовы os.Exit() в функции main()
// - Игнорирует временные файлы сборки (go-build)
//
// Пример нарушения:
//
//	func main() {
//	    os.Exit(1) // Вызовет ошибку анализатора
//	}
//
// Рекомендуемая замена:
//
//	func main() {
//	    log.Fatal("error message") // Корректный вариант
//	}

var NoOsExitAnalyzer = &analysis.Analyzer{
	Name: "noosexit",
	Doc:  "forbid direct os.Exit calls in main function",
	Run:  runNoOsExit,
}

func runNoOsExit(pass *analysis.Pass) (interface{}, error) {
	for _, file := range pass.Files {
		if pass.Pkg.Name() != "main" {
			continue
		}

		filePos := pass.Fset.Position(file.Pos())
		if strings.Contains(filePos.Filename, "go-build") {
			continue
		}

		ast.Inspect(file, func(n ast.Node) bool {
			fn, ok := n.(*ast.FuncDecl)
			if !ok || fn.Name.Name != "main" {
				return true
			}

			ast.Inspect(fn.Body, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}

				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				if ident, ok := sel.X.(*ast.Ident); ok {
					if ident.Name == "os" && sel.Sel.Name == "Exit" {
						pass.Reportf(call.Pos(), "direct call to os.Exit in main function is forbidden")
					}
				}
				return true
			})
			return false
		})
	}
	return nil, nil
}

// GetAnalyzers возвращает все доступные анализаторы.
//
// Включает:
// - Стандартные анализаторы из golang.org/x/tools
// - Кастомный NoOsExitAnalyzer
// - Сторонние анализаторы (errwrap, goone)
func GetAnalyzers() []*analysis.Analyzer {
	return []*analysis.Analyzer{
		printf.Analyzer,     // Проверка форматированных строк
		shadow.Analyzer,     // Обнаружение затенённых переменных
		structtag.Analyzer,  // Валидация тегов структур
		nilfunc.Analyzer,    // Проверка сравнений с nil
		stdmethods.Analyzer, // Проверка стандартных интерфейсов
		unmarshal.Analyzer,  // Проверка методов Unmarshal
		NoOsExitAnalyzer,    // Запрет os.Exit в main
		errwrap.Analyzer,    // Проверка оборачивания ошибок
		goone.Analyzer,      // Обнаружение SQL в циклах
	}
}

// printf.Analyzer: Проверка форматированных строк
//
// Проверяет:
// - Соответствие количества аргументов в Printf-функциях
// - Корректность спецификаторов формата (%s, %d и т.д.)
// - Неиспользуемые аргументы форматирования

// shadow.Analyzer: Обнаружение затенённых переменных
//
// Выявляет:
// - Переменные, перекрывающие объявления из внешних областей видимости
// - Потенциально опасные переопределения переменных

// structtag.Analyzer: Валидация тегов структур
//
// Проверяет:
// - Синтаксическую корректность тегов
// - Соответствие стандартным форматам (json, xml)
// - Дублирование тегов

// nilfunc.Analyzer: Проверка сравнений с nil
//
// Обнаруживает:
// - Бессмысленные сравнения (nil == nil)
// - Избыточные nil-проверки
// - Потенциальные nil-паники

// stdmethods.Analyzer: Проверка стандартных интерфейсов
//
// Валидирует:
// - Корректность реализации String() string
// - Соответствие сигнатурам стандартных интерфейсов
// - Правильность именования методов

// unmarshal.Analyzer: Проверка методов Unmarshal
//
// Анализирует:
// - Сигнатуры методов Unmarshal
// - Соответствие ожидаемым параметрам
// - Возвращаемые значения

// errwrap.Analyzer: Проверка оборачивания ошибок
//
// Требует:
// - Все ошибки должны быть обёрнуты с контекстом
// - Использование fmt.Errorf с %w вместо прямого возврата
// - Проверяет цепочки ошибок

// goone.Analyzer: Обнаружение SQL в циклах
//
// Выявляет:
// - SQL-запросы внутри циклов (проблема N+1)
// - Неоптимальные массовые операции
// - Потенциальные узкие места производительности
