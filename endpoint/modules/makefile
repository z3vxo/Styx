CXX64   = x86_64-w64-mingw32-g++
NASM    = nasm
OBJCOPY = x86_64-w64-mingw32-objcopy

CXXFLAGS  = -Os -fno-asynchronous-unwind-tables -nostdlib
CXXFLAGS += -fno-stack-protector -fno-ident
CXXFLAGS += -fno-exceptions -fno-rtti -mno-red-zone
CXXFLAGS += -fno-builtin -w -masm=intel
CXXFLAGS += -falign-functions=1 -falign-jumps=1 -falign-labels=1
CXXFLAGS += -std=c++20

LDFLAGS = -Wl,-Tlinker.ld,--no-seh,-e,module_start,--gc-sections,-s

OUTDIR  = output
MODULES = $(patsubst %.cc,%,$(wildcard *.cc))
TARGETS = $(addprefix $(OUTDIR)/,$(addsuffix .mod,$(MODULES)))

all: $(TARGETS)

# `make <name>` builds <name>.cc into output/<name>.mod.
$(MODULES): %: $(OUTDIR)/%.mod

$(OUTDIR)/%.mod: %.cc common.o linker.ld common.h | $(OUTDIR)
	@echo "[*] Compiling $<..."
	@$(CXX64) $< common.o -o $*.exe $(CXXFLAGS) $(LDFLAGS)
	@echo "[*] Extracting .text -> $@"
	@$(OBJCOPY) -O binary -j .text $*.exe $@
	@rm -f $*.exe

$(OUTDIR):
	@mkdir -p $@

common.o: common.asm
	@echo "[*] Assembling common stub..."
	@$(NASM) -f win64 common.asm -o common.o

clean:
	@rm -rf $(OUTDIR)
	@rm -f *.o *.exe

.PHONY: all clean $(MODULES)
