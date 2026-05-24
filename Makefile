memo:
	go test "./upd" -memprofile mem.out -bench BenchmarkUPD_FillTemplate
	go test -memprofile mem.out -bench ./upd/fill_template_test.go/BenchmarkUPD_FillTemplate