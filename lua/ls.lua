return function(reader)
    local types = {
        [1] = "directory",
        [2] = "link",
        [3] = "file",
    }

    local entries = {}
    local typeWidth = 4 -- "TYPE"
    local sizeWidth = 4 -- "SIZE"

    while true do
        local nameLen = reader:read4()
        if nameLen == 0xFFFFFFFF then
            break
        end

        local name = reader:readBytes(nameLen)
        local kind = types[reader:read4()] or "?"
        local size = string.format("%.0f", reader:read8())

        if #kind > typeWidth then typeWidth = #kind end
        if #size > sizeWidth then sizeWidth = #size end

        entries[#entries + 1] = { kind = kind, size = size, name = name }
    end

    local format = "%-" .. typeWidth .. "s  %" .. sizeWidth .. "s  %s"
    local lines = { string.format(format, "TYPE", "SIZE", "NAME") }
    for _, entry in ipairs(entries) do
        lines[#lines + 1] = string.format(format, entry.kind, entry.size, entry.name)
    end

    return table.concat(lines, "\n")
end
