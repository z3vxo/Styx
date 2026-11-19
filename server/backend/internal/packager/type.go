package packager




const (
	OUTPUT_NO_PARSE    = 0xc1
	OUTPUT_NEEDS_PARSE = 0xc2
	OUTPUT_SUCCESS     = 0xc3
	OUTPUT_ERROR       = 0xc4


	END_SIG            = 0xFFFFFFFF
)

var CodeToStrings = map[int32]string {
	0xb1: "MODULE: Agent succesfully loaded module",
	10:   "MODULE: Agent already has module loaded.",
	11:   "MODULE: Agent failed to load module.",
}