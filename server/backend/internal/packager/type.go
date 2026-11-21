package packager

const (
	OUTPUT_NO_PARSE    = 0xc1
	OUTPUT_NEEDS_PARSE = 0xc2
	OUTPUT_SUCCESS     = 0xc3
	OUTPUT_ERROR       = 0xc4

	END_SIG = 0xFFFFFFFF
)

var CodeToStrings = map[int32]string{
	0xb1: "MODULE: Agent succesfully loaded module",
	0xb5: "MODULE: Agent succesfully unloaded module",

	0xd1: "CD: Agent succesfully changed directorys",

	10: "MODULE: Agent already has module loaded.",
	11: "MODULE: Agent failed to load module.",
	12: "MODULE: Agent does not have module loaded.",
	13: "MODULE: Agent failed to unload module.",
	20: "CAT: Provided path is not a file.",
}
