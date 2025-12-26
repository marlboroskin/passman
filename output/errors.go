package output

import "github.com/fatih/color"

func PrintError(value any) {
	val, ok := value.(int)
	if ok {
		color.Red("Код ошибки: %d", val)
		return
	}
	strValue, ok := value.(string)
	if ok {
		color.Red(strValue)
		return
	}
	errValue, ok := value.(error)
	if ok {
		color.Red(errValue.Error())
		return
	}
	color.Red("Неизвестный тип ошибки")
}

