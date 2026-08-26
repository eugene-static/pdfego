История версий
===
### Версия v0.1.0
* improved:
    - Полный рефакторинг библиотеки для более гибкого расширения.
    - Один _Core_ теперь может использоваться параллельно разными _Constructor_.
    - Улучшена производительность.
* added:
    - Добавлена возможность использования изображений только для одной итерации. Методы: `Constructor.ReadImage(path, alias string) error`, `Constructor.AddImage(data []byte, alias string) error`.
    - Добавлены методы для установки автора и времени создания документа: `Constructor.SetProducer(producer string)`, `Constructor.SetCreationDate(date time.Time)`.
    - Добавлен метод инициализации наклонного шрифта с псевдонимом _ITALIC_: `SetFontItalic(path string) error`
    - Добавлена возможность задавать свой шрифт по-умолчанию. Методы: `Core.SetDefaultFont(alias string) error`, `Core.SetDefaultFontBold(alias string) error`, `Core.SetDefaultFontItalic(alias string) error`.
    - Добавлен новый объект для _Constructor_: _Iterator_, -- на замену _Repeater_, для более удобной работы с массивами структур. Для _Iterator_ нет необходимости в имплементации интерфейса _Sectioner_. _Iterator_ задается функцией `func NewIterator[T any](collection []T, applier IteratorApplier[T]) *Iterator[T]`. Метод: `Constructor.Iterate(iterator iterator, options ...NodeOptions) *Constructor`
    - Добавлен метод `Constructor.ApplyHeader(applier BlockApplier, options ...NodeOptions) *Constructor` для инлайн-установки _Header_.
    - Добавлен метод `Row.PagesCount(options ...CellOptions) *Row`, который отображает в ячейке количество страниц в документе.
    - Функция `Constructor.ReleaseHeader() *Constructor` теперь так же может вызываться инлайн.
    - Добавлена опция для пагинатора `PaginatorOptions.Placement string` для возможности размещения номера страницы в верхнем колонтитуле.
    - Добавлена опция для пагинатора `PaginatorOptions.Skip uint` для возможности задать количество страниц, которое необходимо пропустить при пагинации.
    - Добавлена опция для ячеек `CellOptions.UnderLine bool`, позволяющая выводить подчеркнутый текст.
    - _Paginator_ получил свой `type PaginatorApplier func(block *Block, number string)`. Использование кастомного пагинатора должно быть изменено соответственно.
    - _Row_ получил свой `type RowApplier func(*Row)`.
* fixed:
    - Словарь ширин ломался, если в тексте был нечитаемый символ с индексом 0.
    - Последовательные индексы группировались в диапазоны максимум из двух элементов.
    - Исправлен баг, когда вызов ячейки, выходящей за диапазон ячеек таблицы, вызывал панику.
    - _Repeater_ работал некорректно, если значение в первой секции было пустым.
### Версия v0.0.2
* fixed:
    - Исправлен неверный подсчет ширины линии текста при автоматическом переносе по пробелам.
* added:
    - `Core.DefaultFontSize()`, `Core.DefaultBorderSize()` теперь экспортируемые.

### Версия v0.0.1
* Релиз