package primitives

const (
	BeginText  Operator = "BT" // Начинает текстовый объект. Все последующие команды до ET относятся к выводу текста.
	TextFont   Operator = "Tf" // Устанавливает текущий шрифт и его кегль.
	TextMatrix Operator = "Tm" // Устанавливает текстовую матрицу для абсолютного позиционирования текста в указанной точке.
	ShowText   Operator = "Tj" // Выводит текстовую строку, используя текущий шрифт и позицию, и сдвигает текстовую позицию на ширину глифа.
	EndText    Operator = "ET" // Завершает текстовый объект.

	StrokeColorRGB  Operator = "RG" // Устанавливает цвет обводки в цветовом пространстве DeviceRGB. Три числа: красный, зелёный, синий (0–1).
	StrokeColorGray Operator = "G"  // Устанавливает цвет обводки в цветовом пространстве DeviceGray. Одно число: серый (0–1).
	StrokeColorCMYK Operator = "K"  // Устанавливает цвет обводки в цветовом пространстве DeviceCMYK. Четыре числа: cyan, magenta, yellow, black (0–1).
	FillColorRGB    Operator = "rg" // Устанавливает цвет заливки в DeviceRGB. Три числа: красный, зелёный, синий (0–1).
	FillColorGray   Operator = "g"  // Устанавливает цвет заливки в DeviceGray. Одно число: серый (0–1).
	FillColorCMYK   Operator = "k"  // Устанавливает цвет заливки в DeviceCMYK. Четыре числа: cyan, magenta, yellow, black (0–1).

	Width  Operator = "w"
	Height Operator = "h"

	MoveTo          Operator = "m"  // Перемещает перо в точку, не рисуя.
	LineTo          Operator = "l"  // Рисует прямую линию от текущей точки до указанной.
	Stroke          Operator = "S"  // Обводит текущий путь (прямоугольник), делает линию видимой.
	CloseAndStroke  Operator = "s"  // Замыкает путь (соединяет последнюю точку с первой) и обводит.
	ClosePath       Operator = "h"  // Замыкает путь, но не обводит.
	Fill            Operator = "f"  // Заливает путь, игнорируя направление.
	FillStroke      Operator = "B"  // Заливает и обводит.
	CloseFillStroke Operator = "b"  // Замыкает, заливает и обводит
	Nop             Operator = "n"  // Завершает путь без отрисовки (используется как «заглушка» или для установки clipping)
	Rectangle       Operator = "re" // Добавляет прямоугольник в текущий путь.

	SaveGraphicsState    Operator = "q"  // Сохраняет текущее графическое состояние (координаты, толщину линий, текущий шрифт, цвет) в специальный стек памяти.
	RestoreGraphicsState Operator = "Q"  // Восстанавливает графическое состояние.
	InvokeNamedXObject   Operator = "Do" // Вызывает выполнение внешнего объекта.
	ConcatenateMatrix    Operator = "cm" // Одновременно сдвигает систему координат в точку (X, Y) и растягивает её до размеров W на H.
)

type Operator string

func (op Operator) Append(dst []byte) []byte {
	return append(dst, op...)
}

func (op Operator) String() string {
	return string(op)
}
