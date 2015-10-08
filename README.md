A solution for dumping parameters in Go by just supplying the function as a
value. Answers http://stackoverflow.com/q/33002720/1643939.

**This is not intended for production use:** parsing stack traces is not the
best way for having stable results over builds. Please take this into
consideration if you are thinking about using this.

### Example usage

Code:

	import "github.com/githubnemo/pdump"

	func Test3(in int) (int, int) {
		defer pdump.PrintInOutputs(Test3)
		return 3, 4
	}

	func main() {
		Test3(42)
	}

Output:

	3,4, = main.Test3(42)


### Requirements

- go 1.5+
