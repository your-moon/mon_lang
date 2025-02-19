package codegen

type MachineTarget string

type Emitter interface {
	Emit()
	starter()
	ending()
}
