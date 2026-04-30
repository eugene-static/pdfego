package template

func (core *Core) addPage() {
	if core.pagesCount > 0 {
		core.pagesContentStream.Reset()

		core.println(
			"endstream",
			"endobj",
		)
	}

	core.pagesCount++

	pageObj := core.getObjNum()

	core.appendOffset()

	core.printInt64(pageObj)
	core.println(
		" 0 obj",
		"<</Type /Page",
		"/Parent 1 0 R",
		"Resources 2 0 R",
	)

	core.print("/Contents 1 0 ")
	core.printInt64(pageObj + 1)
	core.println(
		">>",
		"endobj",
	)

	pageObj = core.getObjNum()

	core.appendOffset()

	core.printInt64(pageObj)
	core.println(" 0 obj")
	core.print("<</Length ")
	core.printInt64(int64(core.pagesContentStream.Len()))
	core.println(
		">>",
		"stream",
	)
	core.pagesContentStream.WriteTo(core.buffer)
	//TODO: обработка ошибок
}
