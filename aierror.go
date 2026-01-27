package aierror

import "fmt"

func Init() {
	fmt.Println("🤖 AI Error Analyzer Enabled")
	fmt.Println("--------------------------------------------------")
}

func Catch() {
	if r := recover(); r != nil {
		Analyze(fmt.Sprintf("%v", r))
	}
}

func Check(err error) {
	if err != nil {
		Analyze(err.Error())
	}
}
