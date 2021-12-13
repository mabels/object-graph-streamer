package main

//c5 "github.com/mabels/c5-envelope/pkg"
import (
	"fmt"

	ogsp "github.com/mabels/object-graph-streamer"
)

type Test struct {
	Yoo int
	Bla int
}

func main() {
	sample := Test{Yoo: 9, Bla: 5}
	out := ""
	jsonC := ogsp.NewJsonCollector(func(o string) { out += o }, &ogsp.JsonProps{})
	ogsp.ObjectGraphStreamer(sample, func(prob ogsp.SVal) {
		jsonC.Append(prob)
	})
	if out != "{\"Bla\":5,\"Yoo\":9}" {
		panic(fmt.Sprintf("Not working%s", out))
	}
	fmt.Println("Ready for production")
}
