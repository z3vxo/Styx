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

MODULE = file-system
OUTDIR = output
TARGET = $(OUTDIR)/$(MODULE).mod

all: file-system

file-system: $(TARGET)

common.o: common.asm
	@ echo "[*] Assembling common stub..."
	@ $(NASM) -f win64 common.asm -o common.o

# Compile the module and save its raw .text as <module>.mod.
$(TARGET): $(MODULE).cc common.o linker.ld common.h
	@ mkdir -p $(OUTDIR)
	@ echo "[*] Compiling $(MODULE)..."
	@ $(CXX64) $< common.o -o $(MODULE).exe $(CXXFLAGS) $(LDFLAGS)
	@ echo "[*] Extracting .text -> $@"
	@ $(OBJCOPY) -O binary -j .text $(MODULE).exe $@
	@ rm -f $(MODULE).exe common.o

clean:
	@ rm -rf $(OUTDIR)
	@ rm -f *.o *.exe

.PHONY: all clean file-system
