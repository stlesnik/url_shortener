/*
Package staticlint — кастомный multichecker, включающий:

1. Стандартные анализаторы (inspect, printf и др.)
2. Все SA-анализаторы staticcheck.io
3. Три других анализатора staticcheck:
  - ST1000
  - S1000
  - QF1001

4. Публичные анализаторы:
  - github.com/gostaticanalysis/nilerr — ловит некорректные сравнения nil и err
  - github.com/gostaticanalysis/forcetypeassert — предупреждает о жестких type assert'ах

5. Кастомный анализатор osexit:
  - Запрещает прямой вызов os.Exit внутри main() пакета main

Запуск:

	go run ./cmd/staticlint ./...
	go build -o staticlint ./cmd/staticlint
	./staticlint ./...
*/
package main
