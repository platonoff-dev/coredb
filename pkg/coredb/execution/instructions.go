package execution

type OpCode uint8

const (
	OpPushConst = iota + 1
	OpPushReg
	OpPopReg
	Swap
	OpDrop
)

type Instruction struct {
	OP OpCode
}
