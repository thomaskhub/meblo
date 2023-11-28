package main

import (
	"fmt"
	"os"
	"time"

	"github.com/thomaskhub/meblo/inputs"
	"github.com/thomaskhub/meblo/logger"
	"github.com/thomaskhub/meblo/outputs"
	"go.uber.org/zap/zapcore"
)

func main() {
	logger.InitLogger(zapcore.ErrorLevel)

	//read the input
	in := inputs.NewInput()
	// defer in.Close()
	fmt.Println(os.Getwd())
	err := in.Open("../../assets/test/test.mp4", false)
	if err != nil {
		panic(err)
	}

	in.Run()

	//get and setup the output
	out := outputs.NewOutput()
	// defer out.Close()

	out.SetDataChannel(*in.GetDataChannel())
	out.SetMetaData(in.GetMetaData())
	out.Open("/tmp/oranges.ts")

	time.Sleep(time.Second * 3)
}
