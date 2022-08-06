package compiler

import (
	"fmt"
	"go/token"
	"go/types"
	"strconv"

	"tinygo.org/x/go-llvm"
)

func (b *builder) createGoAsmWrapper(forwardName string) {
	// Obtain architecture specific things.
	var offsetField types.Type
	var movConstAsm, callAsm, callConstraints string
	switch b.archFamily() {
	case "i386":
		offsetField = types.NewStruct(nil, nil)
		movConstAsm = "movl $$%d, ${0}"
		callAsm = fmt.Sprintf("calll %#v", forwardName)
		callConstraints = "={eax},{eax},~{ebx},~{ecx},~{edx},~{esi},~{edi},~{ebp},~{xmm0},~{xmm1},~{xmm2},~{xmm3},~{xmm4},~{xmm5},~{xmm6},~{xmm7},~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory}"
	case "x86_64":
		offsetField = types.NewStruct(nil, nil)
		movConstAsm = "movq $$%d, ${0}"
		callAsm = fmt.Sprintf("callq %#v", forwardName)
		callConstraints = "={rax},{rax},~{rbx},~{rcx},~{rdx},~{rsi},~{rdi},~{rbp},~{r8},~{r9},~{r10},~{r11},~{r12},~{r13},~{r14},~{r15},~{xmm0},~{xmm1},~{xmm2},~{xmm3},~{xmm4},~{xmm5},~{xmm6},~{xmm7},~{xmm8},~{xmm9},~{xmm10},~{xmm11},~{xmm12},~{xmm13},~{xmm14},~{xmm15},~{xmm16},~{xmm17},~{xmm18},~{xmm19},~{xmm20},~{xmm21},~{xmm22},~{xmm23},~{xmm24},~{xmm25},~{xmm26},~{xmm27},~{xmm28},~{xmm29},~{xmm30},~{xmm31},~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory}"
	case "arm":
		// Storing and reloading R11 because it appears that R11 is reserved on
		// Linux. We store it in the area that's empty anyway (it seems to be
		// used for the return pointer).
		offsetField = types.Typ[types.Uintptr]
		movConstAsm = "mov ${0}, #%d"
		callAsm = fmt.Sprintf("str r11, [sp]\nbl %#v\nldr r11, [sp]", forwardName)
		callConstraints = "={r0},{r0},~{r1},~{r2},~{r3},~{r4},~{r5},~{r6},~{r7},~{r8},~{r9},~{r10},~{r12},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{cpsr},~{memory}"
	case "aarch64":
		offsetField = types.Typ[types.Uintptr]
		movConstAsm = "mov ${0}, #%d"
		callAsm = fmt.Sprintf("bl %#v", forwardName)
		callConstraints = "={x0},{x0},~{x1},~{x2},~{x3},~{x4},~{x5},~{x6},~{x7},~{x8},~{x9},~{x10},~{x11},~{x12},~{x13},~{x14},~{x15},~{x16},~{x17},~{x19},~{x20},~{x21},~{x22},~{x23},~{x24},~{x25},~{x26},~{x27},~{x28},~{lr},~{q0},~{q1},~{q2},~{q3},~{q4},~{q5},~{q6},~{q7},~{q8},~{q9},~{q10},~{q11},~{q12},~{q13},~{q14},~{q15},~{q16},~{q17},~{q18},~{q19},~{q20},~{q21},~{q22},~{q23},~{q24},~{q25},~{q26},~{q27},~{q28},~{q29},~{q30},~{nzcv},~{ffr},~{vg},~{memory}"
		if b.GOOS != "darwin" {
			callConstraints += ",~{x18}"
		}
	default:
		b.addError(b.fn.Pos(), "unknown architecture for Go assembly: "+b.archFamily())
	}

	// Initialize parameters, create entry block, etc.
	b.createFunctionStart(true)

	// Determine the stack layout that's used for the Go ABI.
	// The layout roughly follows this convention (from low to high address):
	//   - empty space (for the return pointer?)
	//   - parameters
	//   - return values
	// More information can be found here (ABI0 is equivalent to the regabi
	// without any integer or floating point registers):
	// https://go.googlesource.com/go/+/refs/heads/master/src/cmd/compile/abi-internal.md
	// We need to use size calculations as used by gc (the regular Go compiler)
	// because that's what the assembly expects. It's usually the same as LLVM,
	// but importantly differs on ARM where int64 is 32-bit aligned in gc but
	// 64-bit aligned according to LLVM (and the ARM AAPCS).
	sizes := types.SizesFor("gc", b.GOARCH)
	var paramValues []llvm.Value
	var paramFields []*types.Var
	for _, param := range b.fn.Params {
		value := b.getValue(param)
		paramValues = append(paramValues, value)
		paramFields = append(paramFields, types.NewField(token.NoPos, nil, param.Name(), param.Type(), false))
	}
	var resultFields []*types.Var
	results := b.fn.Signature.Results()
	for i := 0; i < results.Len(); i++ {
		field := results.At(i)
		resultFields = append(resultFields, types.NewField(token.NoPos, nil, "result_"+strconv.Itoa(i), field.Type(), false))
	}
	paramStruct := types.NewStruct(paramFields, nil)
	stackStructFields := []*types.Var{
		types.NewField(token.NoPos, nil, "offset", offsetField, false),
		types.NewField(token.NoPos, nil, "params", paramStruct, false),
		types.NewField(token.NoPos, nil, "align", types.NewArray(types.Typ[types.Uintptr], 0), false),
		types.NewField(token.NoPos, nil, "results", types.NewStruct(resultFields, nil), false),
	}
	stackStruct := types.NewStruct(stackStructFields, nil)
	stackStructOffsets := sizes.Offsetsof(stackStructFields)

	// Create the alloca.
	// WARNING: we're assuming here that this alloca will be the same as the
	// stack pointer before the call in the assembly. This is not necessarily
	// the case. But it seems to be working.
	// I have tried many other approaches and this appears to be the most
	// reliable way to do it.
	getAllocaSizeType := llvm.FunctionType(b.uintptrType, nil, false)
	getAllocaSizeAsm := llvm.InlineAsm(getAllocaSizeType, fmt.Sprintf(movConstAsm, uint64(sizes.Sizeof(stackStruct))), "=r", true, true, 0, false)
	allocaSize := b.CreateCall(getAllocaSizeType, getAllocaSizeAsm, nil, "")
	alloca := b.CreateArrayAlloca(b.ctx.Int8Type(), allocaSize, "callframe")

	// Store parameters at the top of the stack.
	paramOffsets := sizes.Offsetsof(paramFields)
	for i, param := range paramValues {
		offset := stackStructOffsets[1] + paramOffsets[i]
		b.storeUsingGoLayout(b.fn.Pos(), alloca, sizes, offset, paramFields[i].Type(), param)
	}

	// Call the Go assembly!
	// This is done in inline assembly because ABI0 clobbers more registers than
	// a call would in the C calling convention.
	// The return value exists purely to tell LLVM that this register has been
	// clobbered. The return value from the assembly isn't actually used.
	asmType := llvm.FunctionType(b.i8ptrType, []llvm.Type{b.i8ptrType}, false)
	inlineAsm := llvm.InlineAsm(asmType, callAsm, callConstraints, true, true, 0, false)
	b.CreateCall(asmType, inlineAsm, []llvm.Value{alloca}, "")

	// Read return values.
	resultOffsets := sizes.Offsetsof(resultFields)
	var resultValues []llvm.Value
	for i := range resultFields {
		offset := stackStructOffsets[2] + resultOffsets[i]
		result := b.loadUsingGoLayout(b.fn.Pos(), alloca, sizes, offset, resultFields[i].Type())
		resultValues = append(resultValues, result)
	}

	// Return the resulting value.
	b.createReturn(resultValues)

	// With all the unsafe stuff we do above, it seems better to mark this
	// function as noinline so that optimizations won't interfere too much with
	// it.
	noinline := b.ctx.CreateEnumAttribute(llvm.AttributeKindID("noinline"), 0)
	b.llvmFn.AddFunctionAttr(noinline)
}

func (b *builder) createGoAsmExport(forwardName string) {
	// Obtain some information about the target.
	stackAlignment := uint64(0)
	switch b.archFamily() {
	case "x86_64":
		// Go uses an 8-byte stack alignment. Increase this to a 16-byte
		// alignment.
		stackAlignment = 16
	case "arm":
		// Go appears to use a 4-byte stack alignment. The AAPCS requires an
		// 8-byte alignment, so increase the alignment to 8 bytes.
		stackAlignment = 8
	default:
		// - 386: not sure, need to check this.
		// - arm64: always uses a 16-byte stack alignment
	}

	// Create function that reads incoming parameters from the stack.
	llvmFnType := llvm.FunctionType(b.ctx.VoidType(), []llvm.Type{b.i8ptrType}, false)
	llvmFn := llvm.AddFunction(b.mod, b.info.linkName+"$goasmwrapper", llvmFnType)
	llvmFn.SetLinkage(llvm.InternalLinkage)
	noinline := b.ctx.CreateEnumAttribute(llvm.AttributeKindID("noinline"), 0)
	llvmFn.AddFunctionAttr(noinline)
	if stackAlignment != 0 {
		alignstack := b.ctx.CreateEnumAttribute(llvm.AttributeKindID("alignstack"), stackAlignment)
		llvmFn.AddFunctionAttr(alignstack)
	}
	b.addStandardDeclaredAttributes(llvmFn)
	b.addStandardDefinedAttributes(llvmFn)
	bb := llvm.AddBasicBlock(llvmFn, "entry")
	b.SetInsertPointAtEnd(bb)
	if b.Debug {
		pos := b.program.Fset.Position(b.fn.Pos())
		difunc := b.attachDebugInfoRaw(b.fn, llvmFn, "$goasmwrapper", pos.Filename, pos.Line)
		b.SetCurrentDebugLocation(uint(pos.Line), 0, difunc, llvm.Metadata{})
	}

	// Determine the stack layout that's used for the Go ABI.
	// See createGoAsmWrapper for details.
	sizes := types.SizesFor("gc", b.GOARCH)
	var paramFields []*types.Var
	for _, param := range b.fn.Params {
		paramFields = append(paramFields, types.NewField(token.NoPos, nil, param.Name(), param.Type(), false))
	}
	var resultFields []*types.Var
	for i := 0; i < b.fn.Signature.Results().Len(); i++ {
		resultFields = append(resultFields, b.fn.Signature.Results().At(i))
	}
	paramStruct := types.NewStruct(paramFields, nil)
	stackStructFields := []*types.Var{
		types.NewField(token.NoPos, nil, "offset", types.Typ[types.Uintptr], false),
		types.NewField(token.NoPos, nil, "params", paramStruct, false),
		types.NewField(token.NoPos, nil, "align", types.NewArray(types.Typ[types.Uintptr], 0), false),
		types.NewField(token.NoPos, nil, "results", types.NewStruct(resultFields, nil), false),
	}
	stackStructOffsets := sizes.Offsetsof(stackStructFields)

	// Read parameters from the stack.
	sp := llvmFn.Param(0)
	var params []llvm.Value
	paramOffsets := sizes.Offsetsof(paramFields)
	for i, param := range paramFields {
		offset := stackStructOffsets[1] + paramOffsets[i]
		value := b.loadUsingGoLayout(b.fn.Pos(), sp, sizes, offset, param.Type())
		params = append(params, value)
	}

	// Call the Go function!
	params = append(params, llvm.ConstNull(b.i8ptrType)) // context
	resultValue := b.createCall(b.llvmFnType, b.llvmFn, params, "result")

	// Split the result value into a slice, to match resultFields.
	var resultValues []llvm.Value
	if len(resultFields) == 1 {
		resultValues = []llvm.Value{resultValue}
	} else if len(resultFields) > 1 {
		for i := range resultFields {
			value := b.CreateExtractValue(resultValue, i, "")
			resultValues = append(resultValues, value)
		}
	}

	// Store the result in the stack space reserved by the Go assembly.
	resultOffsets := sizes.Offsetsof(resultFields)
	for i, result := range resultFields {
		offset := stackStructOffsets[2] + resultOffsets[i]
		b.storeUsingGoLayout(b.fn.Pos(), sp, sizes, offset, result.Type(), resultValues[i])
	}

	// Values are returned by passing them in a special way on the stack, not
	// via a conventional return.
	b.CreateRetVoid()

	// TODO: use llvm.sponentry when available (ARM and AArch64 as of LLVM 15).
	b.createGoAsmSPForward(forwardName, llvmFnType, llvmFn)
}

// Create a stub function that captures the stack pointer and passes it to
// another function.
// We should really be using llvm.sponentry when available, or even port it to
// new architectures as needed.
func (b *builder) createGoAsmSPForward(forwardName string, forwardFunctionType llvm.Type, forwardFunction llvm.Value) {
	// Create function that forwards the stack pointer.
	llvmFnType := llvm.FunctionType(b.ctx.VoidType(), nil, false)
	llvmFn := llvm.AddFunction(b.mod, forwardName, llvmFnType)
	b.addStandardDeclaredAttributes(llvmFn)
	b.addStandardDefinedAttributes(llvmFn)
	bb := llvm.AddBasicBlock(llvmFn, "entry")
	b.Builder.SetInsertPointAtEnd(bb)
	if b.Debug {
		pos := b.program.Fset.Position(b.fn.Pos())
		b.SetCurrentDebugLocation(uint(pos.Line), 0, b.difunc, llvm.Metadata{})
	}
	asmType := llvm.FunctionType(b.i8ptrType, nil, false)
	var asmString string
	switch b.archFamily() {
	case "x86_64":
		asmString = "mov %rsp, $0"
	case "arm":
		asmString = "mov r0, sp"
	case "aarch64":
		asmString = "mov x0, sp"
	default:
		b.addError(b.fn.Pos(), "cannot wrap Go assembly: unknown architecture")
		b.CreateUnreachable()
		return
	}
	asm := llvm.InlineAsm(asmType, asmString, "=r", false, false, 0, false)
	sp := b.CreateCall(asmType, asm, nil, "sp")
	call := b.CreateCall(forwardFunctionType, forwardFunction, []llvm.Value{sp}, "")
	call.SetTailCall(true)
	b.CreateRetVoid()

	// Make sure no prologue/epilogue is created (that would change the stack
	// pointer).
	naked := b.ctx.CreateEnumAttribute(llvm.AttributeKindID("naked"), 0)
	llvmFn.AddFunctionAttr(naked)
}

// Load a value from the given pointer with the given offset, assuming the
// memory layout in the sizes parameter. The ptr must be of type *i8.
func (b *builder) loadUsingGoLayout(pos token.Pos, ptr llvm.Value, sizes types.Sizes, offset int64, typ types.Type) llvm.Value {
	typ = typ.Underlying()
	switch typ := typ.(type) {
	case *types.Basic, *types.Pointer, *types.Slice:
		gep := b.CreateGEP(b.ctx.Int8Type(), ptr, []llvm.Value{llvm.ConstInt(b.ctx.Int32Type(), uint64(offset), false)}, "")
		valueType := b.getLLVMType(typ)
		bitcast := b.CreateBitCast(gep, llvm.PointerType(valueType, 0), "")
		return b.CreateLoad(valueType, bitcast, "")
	default:
		b.addError(pos, "todo: unknown type to load: "+typ.String())
		return llvm.Undef(b.getLLVMType(typ))
	}
}

// Store a value at the address given by ptr with the given offset, assuming the
// memory layout in the sizes parameter. The ptr must be of type *i8.
func (b *builder) storeUsingGoLayout(pos token.Pos, ptr llvm.Value, sizes types.Sizes, offset int64, typ types.Type, value llvm.Value) {
	typ = typ.Underlying()
	switch typ := typ.(type) {
	case *types.Basic, *types.Pointer, *types.Slice:
		gep := b.CreateGEP(b.ctx.Int8Type(), ptr, []llvm.Value{llvm.ConstInt(b.ctx.Int32Type(), uint64(offset), false)}, "")
		bitcast := b.CreateBitCast(gep, llvm.PointerType(b.getLLVMType(typ), 0), "")
		b.CreateStore(value, bitcast)
	default:
		b.addError(pos, "todo: unknown type to store: "+typ.String())
	}
}
