; ModuleID = 'goasm.go'
source_filename = "goasm.go"
target datalayout = "e-m:e-p270:32:32-p271:32:32-p272:64:64-i64:64-f80:128-n8:16:32:64-S128"
target triple = "x86_64-unknown-linux"

@main.asmGlobalExport = hidden global i32 0, align 4

@__GoABI0_main.asmGlobalExport = alias i32, ptr @main.asmGlobalExport

declare noalias nonnull ptr @runtime.alloc(i64, ptr, ptr) #0

; Function Attrs: nounwind uwtable(sync)
define hidden void @main.init(ptr %context) unnamed_addr #1 {
entry:
  ret void
}

; Function Attrs: noinline nounwind uwtable(sync)
define hidden double @main.AsmSqrt(double %x, ptr %context) unnamed_addr #2 {
entry:
  %0 = call i64 asm sideeffect alignstack "movq $$16, ${0}", "=r"() #5
  %callframe = alloca i8, i64 %0, align 8
  store double %x, ptr %callframe, align 8
  %1 = call ptr asm sideeffect alignstack "callq \22__GoABI0_main.AsmSqrt\22", "={rax},{rax},~{rbx},~{rcx},~{rdx},~{rsi},~{rdi},~{rbp},~{r8},~{r9},~{r10},~{r11},~{r12},~{r13},~{r14},~{r15},~{xmm0},~{xmm1},~{xmm2},~{xmm3},~{xmm4},~{xmm5},~{xmm6},~{xmm7},~{xmm8},~{xmm9},~{xmm10},~{xmm11},~{xmm12},~{xmm13},~{xmm14},~{xmm15},~{xmm16},~{xmm17},~{xmm18},~{xmm19},~{xmm20},~{xmm21},~{xmm22},~{xmm23},~{xmm24},~{xmm25},~{xmm26},~{xmm27},~{xmm28},~{xmm29},~{xmm30},~{xmm31},~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory}"(ptr nonnull %callframe) #5
  %2 = getelementptr i8, ptr %callframe, i64 8
  %3 = load double, ptr %2, align 8
  ret double %3
}

; Function Attrs: noinline nounwind uwtable(sync)
define hidden double @main.AsmAdd(double %x, double %y, ptr %context) unnamed_addr #2 {
entry:
  %0 = call i64 asm sideeffect alignstack "movq $$24, ${0}", "=r"() #5
  %callframe = alloca i8, i64 %0, align 8
  store double %x, ptr %callframe, align 8
  %1 = getelementptr i8, ptr %callframe, i64 8
  store double %y, ptr %1, align 8
  %2 = call ptr asm sideeffect alignstack "callq \22__GoABI0_main.AsmAdd\22", "={rax},{rax},~{rbx},~{rcx},~{rdx},~{rsi},~{rdi},~{rbp},~{r8},~{r9},~{r10},~{r11},~{r12},~{r13},~{r14},~{r15},~{xmm0},~{xmm1},~{xmm2},~{xmm3},~{xmm4},~{xmm5},~{xmm6},~{xmm7},~{xmm8},~{xmm9},~{xmm10},~{xmm11},~{xmm12},~{xmm13},~{xmm14},~{xmm15},~{xmm16},~{xmm17},~{xmm18},~{xmm19},~{xmm20},~{xmm21},~{xmm22},~{xmm23},~{xmm24},~{xmm25},~{xmm26},~{xmm27},~{xmm28},~{xmm29},~{xmm30},~{xmm31},~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory}"(ptr nonnull %callframe) #5
  %3 = getelementptr i8, ptr %callframe, i64 16
  %4 = load double, ptr %3, align 8
  ret double %4
}

; Function Attrs: noinline nounwind uwtable(sync)
define hidden { i64, double } @main.AsmFoo(double %x, ptr %context) unnamed_addr #2 {
entry:
  %0 = call i64 asm sideeffect alignstack "movq $$24, ${0}", "=r"() #5
  %callframe = alloca i8, i64 %0, align 8
  store double %x, ptr %callframe, align 8
  %1 = call ptr asm sideeffect alignstack "callq \22__GoABI0_main.AsmFoo\22", "={rax},{rax},~{rbx},~{rcx},~{rdx},~{rsi},~{rdi},~{rbp},~{r8},~{r9},~{r10},~{r11},~{r12},~{r13},~{r14},~{r15},~{xmm0},~{xmm1},~{xmm2},~{xmm3},~{xmm4},~{xmm5},~{xmm6},~{xmm7},~{xmm8},~{xmm9},~{xmm10},~{xmm11},~{xmm12},~{xmm13},~{xmm14},~{xmm15},~{xmm16},~{xmm17},~{xmm18},~{xmm19},~{xmm20},~{xmm21},~{xmm22},~{xmm23},~{xmm24},~{xmm25},~{xmm26},~{xmm27},~{xmm28},~{xmm29},~{xmm30},~{xmm31},~{fpsr},~{fpcr},~{flags},~{dirflag},~{memory}"(ptr nonnull %callframe) #5
  %2 = getelementptr i8, ptr %callframe, i64 8
  %3 = load i64, ptr %2, align 8
  %4 = getelementptr i8, ptr %callframe, i64 16
  %5 = load double, ptr %4, align 8
  %6 = insertvalue { i64, double } zeroinitializer, i64 %3, 0
  %7 = insertvalue { i64, double } %6, double %5, 1
  ret { i64, double } %7
}

; Function Attrs: nounwind uwtable(sync)
define hidden double @main.asmExport(double %x, ptr %context) unnamed_addr #1 {
entry:
  ret double 0.000000e+00
}

; Function Attrs: noinline nounwind alignstack(16) uwtable(sync)
define internal void @"main.asmExport$goasmwrapper"(ptr %0) #3 {
entry:
  %1 = getelementptr i8, ptr %0, i64 8
  %2 = load double, ptr %1, align 8
  %result = call double @main.asmExport(double %2, ptr null)
  %3 = getelementptr i8, ptr %0, i64 16
  store double %result, ptr %3, align 8
  ret void
}

; Function Attrs: naked nounwind uwtable(sync)
define void @__GoABI0_main.asmExport() #4 {
entry:
  %sp = call ptr asm "mov %rsp, $0", "=r"() #5
  tail call void @"main.asmExport$goasmwrapper"(ptr %sp)
  ret void
}

attributes #0 = { "target-features"="+cx8,+fxsr,+mmx,+sse,+sse2,+x87" }
attributes #1 = { nounwind uwtable(sync) "target-features"="+cx8,+fxsr,+mmx,+sse,+sse2,+x87" }
attributes #2 = { noinline nounwind uwtable(sync) "target-features"="+cx8,+fxsr,+mmx,+sse,+sse2,+x87" }
attributes #3 = { noinline nounwind alignstack=16 uwtable(sync) "target-features"="+cx8,+fxsr,+mmx,+sse,+sse2,+x87" }
attributes #4 = { naked nounwind uwtable(sync) "target-features"="+cx8,+fxsr,+mmx,+sse,+sse2,+x87" }
attributes #5 = { nounwind }
