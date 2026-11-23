-- do {
-- 		if (pStrcmp(fData.cFileName, dot1) == 0 || pStrcmp(fData.cFileName, dot2) == 0) { continue; }
--
-- 		if (fData.dwFileAttributes & FILE_ATTRIBUTE_DIRECTORY) { EntryType = 1; }
-- 		if (fData.dwFileAttributes & FILE_ATTRIBUTE_REPARSE_POINT) { EntryType = 3; }
-- 		else { EntryType = 3; }
--
-- 		INT32 Len = (INT32)pStrlen(fData.cFileName);
-- 		mod->ModulePack4(Len);
-- 		mod->ModulePackBytes((PBYTE)fData.cFileName, Len);
-- 		mod->ModulePack4(EntryType);
-- 		ULONGLONG size = ((ULONGLONG)fData.nFileSizeHigh << 32) | fData.nFileSizeLow;
-- 		mod->ModulePack8((INT64)size);
--
-- 	} while (fnFindNextFileA(hFind, &fData));
-- 	mod->ModulePack4(END_SIG);


return function(reader)
    local lines = {}
    local types = {
        [1] = "directory",
        [2] = "link",
        [3] = "file",
    }
    while true do
        local StrLen = reader:read4()
        if StrLen == 0xFFFFFFFF then
            break
        end

        local fileName = reader:readBytes(StrLen)
        local FileType = reader:read4()
        local fileSize = reader:read8()
        local typeChar = types[FileType] or "?"

        lines[#lines + 1] = string.format(
            "%srw-r--r-- %12.0f %s",
            typeChar,
            fileSize,
            fileName
        )
    end
    return table.concat(lines, "\n")
end
